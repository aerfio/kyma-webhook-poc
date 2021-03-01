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
		ServiceAcccountAllowList []string
		NamespaceDenyList        []string
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
			name: "should deny action performed by denied sa in denied namespace",
			fields: fields{
				ServiceAcccountAllowList: []string{},
				NamespaceDenyList:        []string{"kyma-system", "some-other-namespace"},
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
		{
			name: "should allow action performed by denied sa in allowed namespace",
			fields: fields{
				ServiceAcccountAllowList: []string{},
				NamespaceDenyList:        []string{"kyma-system", "sth", "doesntmatter"},
			},
			args: args{
				req: admission.Request{
					AdmissionRequest: admissionv1.AdmissionRequest{
						Namespace: "some-other-namespace",
						UserInfo: authenticationv1.UserInfo{
							Username: "denied-sa",
						},
					},
				},
			},
			want: admission.Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					Allowed: true,
				},
			},
		},
		{
			name: "should allow action performed by allowed sa in denied namespace",
			fields: fields{
				ServiceAcccountAllowList: []string{"allowed-sa"},
				NamespaceDenyList:        []string{"kyma-system"},
			},
			args: args{
				req: admission.Request{
					AdmissionRequest: admissionv1.AdmissionRequest{
						Namespace: "kyma-system",
						UserInfo: authenticationv1.UserInfo{
							Username: "allowed-sa",
						},
					},
				},
			},
			want: admission.Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					Allowed: true,
				},
			},
		},
		{
			name: "should deny action performed by denied sa in clusterscope",
			fields: fields{
				ServiceAcccountAllowList: []string{},
				NamespaceDenyList:        []string{"", "some-namespace"},
			},
			args: args{
				req: admission.Request{
					AdmissionRequest: admissionv1.AdmissionRequest{
						Namespace: "",
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
		{
			name:   "should not crash with no sa-list and ns-list",
			fields: fields{},
			args: args{
				req: admission.Request{
					AdmissionRequest: admissionv1.AdmissionRequest{
						Namespace: "some-ns",
						UserInfo: authenticationv1.UserInfo{
							Username: "some-sa",
						},
					},
				},
			},
			want: admission.Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					Allowed: true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator(logr.DiscardLogger{}, tt.fields.NamespaceDenyList, tt.fields.ServiceAcccountAllowList)

			got := v.Handle(context.Background(), tt.args.req)

			assert.Equal(t, got.Allowed, tt.want.Allowed)
			assert.NotNil(t, got.Result)

			// this is not a mutating webhook, do not add ANY mutations here
			assert.Nil(t, got.PatchType)
			assert.Nil(t, got.Patches)

			if !tt.want.Allowed {
				assert.NotEmpty(t, got.Result.Reason)
			}
		})
	}
}

func Test_extractNsFromUsername(t *testing.T) {

	tests := []struct {
		name    string
		sa      string
		want    string
		wantErr bool
	}{
		{
			name:    "should deny string with wrong prefix",
			sa:      "wrong-sa-string",
			want:    "",
			wantErr: true,
		},
		{
			name:    "should deny wrong sa string, with correct prefix",
			sa:      "system:serviceaccount:something",
			want:    "",
			wantErr: true,
		},
		{
			name:    "should correctly extract namespace",
			sa:      "system:serviceaccount:kyma-system:test-deny",
			want:    "kyma-system",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractNsFromUsername(tt.sa)
			if tt.wantErr {
				assert.NotNil(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_isUsernameStr(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     bool
	}{
		{
			name:     "should return false for uncorrect sa string",
			username: "test",
			want:     false,
		},

		{
			name:     "should correctly parse correct sa string",
			username: "system:serviceaccount:default:test-deny",
			want:     true,
		},
		{
			name:     "should deny group string",
			username: "logged.via.mail@email.com",
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isUsernameStr(tt.username))
		})
	}
}
