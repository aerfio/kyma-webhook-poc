package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type PodValidator struct {
	Client  client.Client
	Decoder *admission.Decoder
}

func (v *PodValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}

	err := v.Decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	data, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	fmt.Printf("%s\n", string(data))

	if req.UserInfo.Username == "system:serviceaccount:default:test-deny" {
		return admission.Denied("this sa is always denied")
	}

	key := "foo"
	label, found := pod.Labels[key]
	if !found {
		return admission.Denied(fmt.Sprintf("missing label %s", key))
	}
	if label != "bar" {
		return admission.Denied(fmt.Sprintf("label %s did not have value %q", key, "foo"))
	}

	return admission.Allowed("")
}
