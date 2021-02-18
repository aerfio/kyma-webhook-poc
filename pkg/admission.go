package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type Validator struct {
	ServiceAccountDenyList []string
	NamespaceDenyList      []string    // "" means clusterwide
	Log                    logr.Logger `envconfig:"-"`
}

func (v *Validator) Handle(_ context.Context, req admission.Request) admission.Response {
	if lg := v.Log.V(1); lg.Enabled() { // V(1) == debug
		marshalledReq, err := json.Marshal(req)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		lg.Info(string(marshalledReq))
	}

	fmt.Printf("%s\n", "----------")
	fmt.Printf("%s\n", req.Namespace)
	fmt.Printf("%v\n", v.isDeniedNamespace(req.Namespace))
	fmt.Printf("%s\n", req.UserInfo.Username)
	fmt.Printf("%v\n", v.isDeniedServiceAccount(req.UserInfo.Username))
	fmt.Printf("%s\n", "----------")
	if v.isDeniedNamespace(req.Namespace) && v.isDeniedServiceAccount(req.UserInfo.Username) {
		scopeErrMsg := fmt.Sprintf("in namespace %s", req.Namespace)
		if req.Namespace == "" {
			scopeErrMsg = "in clusterwide scope"
		}
		return admission.Denied(fmt.Sprintf("ServiceAccount %s is denied to perform action %s", req.UserInfo.Username, scopeErrMsg))
	}

	return admission.Allowed("")
}

func (v Validator) isDeniedServiceAccount(sa string) bool {
	return contains(v.ServiceAccountDenyList, sa)
}

func (v Validator) isDeniedNamespace(ns string) bool {
	return contains(v.NamespaceDenyList, ns)
}

func contains(slice []string, element string) bool {
	// yeah, why create such a function in stdlib, who would need it? /s
	fmt.Printf("%+v\n", slice)
	fmt.Printf("%q\n", element)
	fmt.Printf("%s", element)
	for _, s := range slice {
		if s == element {
			return true
		}
	}
	return false
}
