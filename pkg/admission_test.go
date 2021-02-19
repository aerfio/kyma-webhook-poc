package pkg

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	admissionv1 "k8s.io/api/admission/v1"

	authenticationv1 "k8s.io/api/authentication/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func Test_contains(t *testing.T) {
	type args struct {
		slice   []string
		element string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "returns true if slice includes particular string",
			args: args{
				slice:   []string{"1", "2", "3"},
				element: "1",
			},
			want: true,
		},
		{
			name: "returns true if slice includes particular string more than 1 time",
			args: args{
				slice:   []string{"1", "1", "3"},
				element: "1",
			},
			want: true,
		},
		{
			name: "returns false if slice doesnt include particular string",
			args: args{
				slice:   []string{"1", "2", "3"},
				element: "4",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t,
				contains(tt.args.slice, tt.args.element),
				tt.want)
		})
	}
}

func TestNewValidator(t *testing.T) {
	type args struct {
		namespaceDenyList []string
	}
	tests := []struct {
		name string
		args args
		want *Validator
	}{
		{
			name: "correctly parses env arguments",
			args: args{namespaceDenyList: []string{`""`, "kyma-system"}},
			want: &Validator{
				NamespaceDenyList: []string{"", "kyma-system"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewValidator(logr.DiscardLogger{}, tt.args.namespaceDenyList, []string{})
			assert.Equal(t, got.NamespaceDenyList, tt.want.NamespaceDenyList)
		})
	}
}

func TestValidator_Handle(t *testing.T) {
	type fields struct {
		ServiceAccountDenyList []string
		NamespaceDenyList      []string
	}
	type args struct {
		req admission.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   admission.Response
	}{
		{
			name: "sth",
			fields: fields{
				ServiceAccountDenyList: []string{"denied-sa"},
				NamespaceDenyList:      []string{"kyma-system"},
			},
			args: args{
				req: admission.Request{
					AdmissionRequest: admissionv1.AdmissionRequest{
						Namespace: "kyma-system",
						UserInfo: authenticationv1.UserInfo{
							Username: "denied-sa",
						},
					},
				},
			},
			want: admission.Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					Allowed: false,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Validator{
				ServiceAccountDenyList: tt.fields.ServiceAccountDenyList,
				NamespaceDenyList:      tt.fields.NamespaceDenyList,
				Log:                    logr.DiscardLogger{}, // do not log in tests
			}
			got := v.Handle(context.Background(), tt.args.req)
			// msg := got.Result.Message
			// assert.Equal(t, msg, "tst")

			assert.Equal(t, got.Allowed, tt.want.Allowed)
		})
	}
}
