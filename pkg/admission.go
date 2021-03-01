package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type Validator struct {
	ServiceAccountAllowList []string
	NamespaceDenyList       []string    // "" means clusterwide
	Log                     logr.Logger `envconfig:"-"`
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
		ServiceAccountAllowList: serviceAccountDenyList,
		NamespaceDenyList:       nsList,
		Log:                     logger,
	}
}

func (v *Validator) Handle(ctx context.Context, req admission.Request) admission.Response {
	if lg := v.Log.V(1); lg.Enabled() { // V(1) == debug
		marshalledReq, err := json.Marshal(req)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, errors.Wrapf(err, "while marshalling request from %s", req.UserInfo.Username))
		}
		lg.Info(string(marshalledReq))
	}

	return v.handle(ctx, req)
}

func (v *Validator) handle(_ context.Context, req admission.Request) admission.Response {
	username := req.UserInfo.Username
	if isUsernameStr(username) {
		ns, err := extractNsFromUsername(username)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		if ns == req.Namespace {
			return admission.Allowed("")
		}
	}

	if v.isDeniedNamespace(req.Namespace) && !v.isAllowedServiceAccount(username) {
		scopeErrMsg := fmt.Sprintf("in namespace %s", req.Namespace)
		if req.Namespace == "" {
			scopeErrMsg = "in clusterwide scope"
		}
		return admission.Denied(fmt.Sprintf("ServiceAccount %s is denied to perform action %s", username, scopeErrMsg))
	}

	return admission.Allowed("")
}

func (v Validator) isAllowedServiceAccount(sa string) bool {
	return !contains(v.ServiceAccountAllowList, sa)
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

func isUsernameStr(username string) bool {
	prefix := "system:serviceaccount:"
	return strings.HasPrefix(username, prefix) && len(strings.Split(username, ":")) == 4
}

func extractNsFromUsername(sa string) (string, error) {
	prefix := "system:serviceaccount:"
	if !strings.HasPrefix(sa, prefix) {
		return "", fmt.Errorf("expected %s to be prefixed with %s", sa, prefix)
	}

	data := strings.Split(strings.TrimPrefix(sa, prefix), ":")
	if len(data) != 2 {
		return "", fmt.Errorf("expected %s to have 3 ':' in it", sa)
	}

	return data[0], nil
}
