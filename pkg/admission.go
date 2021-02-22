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
	NamespaceDenyList      []string // "" means clusterwide
	Log                    logr.Logger `envconfig:"-"`
}

func NewValidator(logger logr.Logger, namespaceDenyList, serviceAccountDenyList []string) *Validator {
	nsList := make([]string, 0, len(namespaceDenyList))
	for _, ns := range namespaceDenyList {
		if ns == `""` { // this case is needed to correctly parse clusterwide scope from envconfig
			nsList = append(nsList, "")
		} else {
			nsList = append(nsList, ns)
		}
	}

	return &Validator{
		ServiceAccountDenyList: serviceAccountDenyList,
		NamespaceDenyList:      nsList,
		Log:                    logger,
	}
}

func (v *Validator) Handle(_ context.Context, req admission.Request) admission.Response {
	if lg := v.Log.V(1); lg.Enabled() { // V(1) == debug
		marshalledReq, err := json.Marshal(req)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		lg.Info(string(marshalledReq))
	}

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
	for _, s := range slice {
		if s == element {
			return true
		}
	}
	return false
}
