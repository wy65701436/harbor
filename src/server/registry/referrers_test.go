package registry

//
//import (
//	"net/http"
//	"net/http/httptest"
//	"testing"
//)
//
//func TestReferrersHandler(t *testing.T) {
//	// Create a new request
//	req, err := http.NewRequest("GET", "/repositories/myrepo/tags/mytag", nil)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	// Create a new ResponseRecorder to record the response
//	rr := httptest.NewRecorder()
//
//	// Create a new context and add the repository and reference parameters to it
//	ctx := req.Context()
//	ctx = router.WithParam(ctx, "splat", "myrepo")
//	ctx = router.WithParam(ctx, "reference", "mytag")
//
//	// Call the ServeHTTP function of the referrersHandler
//	handler := &referrersHandler{}
//	handler.ServeHTTP(rr, req.WithContext(ctx))
//
//	// Check the status code
//	if status := rr.Code; status != http.StatusOK {
//		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
//	}
//
//	// Check the response header
//	if header := rr.Header().Get("Content-Type"); header != "application/json; charset=utf-8" {
//		t.Errorf("handler returned wrong header: got %v want %v", header, "application/json; charset=utf-8")
//	}
//
//	// Check the response body
//	expected := `{"Link":"","X-Total-Count":0,"body":{"schemaVersion":2,"mediaType":"application/vnd.oci.image.index.v1+json","manifests":[]}}`
//	if rr.Body.String() != expected {
//		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
//	}
//}
