package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/admission/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleUploadCertToWaf(t *testing.T) {
	admissionReview, requestId := getAdmissionReview()
	createOrUpdateCertificate = func(secret apiv1.Secret) (*string, error) {
		certId := "12345"
		return &certId, nil
	}

	request, err := http.NewRequest("PUT", "/upload-cert-to-waf", bytes.NewReader(admissionReview))

	assert.Nil(t, err)

	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleUploadCertToWaf)

	handler.ServeHTTP(responseRecorder, request)

	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	expectedBody := fmt.Sprintf(`{"kind":"UPDATE",`+
		`"response":{"uid":"%s","allowed":true,"patch":"W3sib3AiOiJhZGQiLCJwYXRoIjoiL21ldGFkYXRhL2Fubm90YXRp`+
		`b25zIiwidmFsdWUiOnsid2FmLWNlcnQtdXBsb2FkZXIuaWl0cy50ZWNoL2NlcnQtd2FmLWlkIjoiMTIzNDUiLCJ3YWYtY2VydC11`+
		`cGxvYWRlci5paXRzLnRlY2gvd2FmLWRvbWFpbi1pZCI6IjQ1NjU2MTY1ZGE2NTQ1NiJ9fV0=",`+
		`"patchType":"JSONPatch"}}`, requestId)
	assert.Equal(t, expectedBody, responseRecorder.Body.String())
}

func TestHandleUploadCertToWaf_rejectDueToAnError(t *testing.T) {
	admissionReview, requestId := getAdmissionReview()
	createOrUpdateCertificate = func(secret apiv1.Secret) (*string, error) {
		return nil, errors.New("any error")
	}

	request, err := http.NewRequest("PUT", "/upload-cert-to-waf", bytes.NewReader(admissionReview))

	assert.Nil(t, err)

	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleUploadCertToWaf)

	handler.ServeHTTP(responseRecorder, request)

	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	expectedBody := fmt.Sprintf(`{"kind":"UPDATE","response":{"uid":"%s","allowed":false}}`, requestId)
	assert.Equal(t, expectedBody, responseRecorder.Body.String())
}

func TestHandleUploadCertToWaf_invalidBody(t *testing.T) {
	admissionReview := getInvalidAdmissionReview()

	createOrUpdateCertificate = func(secret apiv1.Secret) (*string, error) {
		return nil, nil
	}

	request, err := http.NewRequest("PUT", "/upload-cert-to-waf", bytes.NewReader(admissionReview))

	assert.Nil(t, err)

	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleUploadCertToWaf)

	handler.ServeHTTP(responseRecorder, request)

	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
}

func getInvalidAdmissionReview() []byte {
	admissionReview := v1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "UPDATE",
			APIVersion: "invalid",
		},
		Response: nil,
	}
	marshalled, _ := json.Marshal(admissionReview)
	return marshalled
}

func getAdmissionReview() ([]byte, types.UID) {
	secret := apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			ResourceVersion: "version1",
			Annotations: map[string]string{
				"waf-cert-uploader.iits.tech/waf-domain-id": "45656165da65456",
			},
		},
		Data: map[string][]byte{
			"tls.crt": []byte("any cert"),
			"tls.key": []byte("any private key"),
		},
	}
	requestId := uuid.NewUUID()
	secretMarshalled, _ := json.Marshal(secret)
	admissionRequest := v1.AdmissionRequest{
		UID:    requestId,
		Kind:   metav1.GroupVersionKind{Kind: "UPDATE"},
		Object: runtime.RawExtension{Raw: secretMarshalled},
	}
	admissionReview := v1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind: "UPDATE",
		},
		Request:  &admissionRequest,
		Response: nil,
	}
	marshalled, _ := json.Marshal(admissionReview)
	return marshalled, requestId
}
