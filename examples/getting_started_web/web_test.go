/* *********************************************************************
 * This Original Work is copyright of 51 Degrees Mobile Experts Limited.
 * Copyright 2025 51 Degrees Mobile Experts Limited, Davidson House,
 * Forbury Square, Reading, Berkshire, United Kingdom RG1 3EU.
 *
 * This Original Work is licensed under the European Union Public Licence (EUPL)
 * v.1.2 and is subject to its terms as set out below.
 *
 * If a copy of the EUPL was not distributed with this file, You can obtain
 * one at https://opensource.org/licenses/EUPL-1.2.
 *
 * The 'Compatible Licences' set out in the Appendix to the EUPL (as may be
 * amended by the European Commission) shall be deemed incompatible for
 * the purposes of the Work and the provisions of the compatibility
 * clause in Article 5 of the EUPL shall not apply.
 *
 * If using the Work as, or as part of, a network application, by
 * including the attribution notice(s) required under Article 5 of the EUPL
 * in the end user terms of the application under an appropriate heading,
 * such notice(s) shall fulfill the requirements of that article.
 * ********************************************************************* */

package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/51Degrees/ip-intelligence-go/v4/ipi_interop"
	"github.com/51Degrees/ip-intelligence-go/v4/ipi_onpremise"
)

var testHandler *ipiHandler

func TestMain(m *testing.M) {
	if runtime.GOOS == "windows" {
		log.Println("InMemory profile currently fails on Windows, skipping")
		os.Exit(0)
	}

	dataFile := os.Getenv("DATA_FILE")
	if dataFile == "" {
		dataFile = "../51Degrees-EnterpriseIpiV41.ipi"
	}

	config := ipi_interop.NewConfigIpi(ipi_interop.LowMemory)
	engine, err := ipi_onpremise.New(
		ipi_onpremise.WithConfigIpi(config),
		ipi_onpremise.WithDataFile(dataFile),
		ipi_onpremise.WithAutoUpdate(false),
	)
	if err != nil {
		log.Printf("Skipping web handler tests: IP Intelligence engine init failed: %v", err)
		os.Exit(0)
	}

	testHandler = &ipiHandler{engine: engine}

	code := m.Run()

	engine.Stop()
	os.Exit(code)
}

// TestHandler_HTMLResponse verifies that the handler returns a well-formed
// HTML page containing IP intelligence results.
func TestHandler_HTMLResponse(t *testing.T) {
	if testHandler == nil {
		t.Skip("engine not initialised; set DATA_FILE")
	}

	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}
	r.Header.Set("X-Forwarded-For", "185.28.167.77")

	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, r)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected HTTP 200, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		t.Errorf("Expected Content-Type text/html, got %q", contentType)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "IP Intelligence") {
		t.Error("Expected page to contain 'IP Intelligence'")
	}
	if !strings.Contains(body, "185.28.167.77") {
		t.Error("Expected page to contain the looked-up IP address")
	}
}

// TestHandler_ClientIPQueryParam verifies that the client-ip query parameter
// is used for the lookup.
func TestHandler_ClientIPQueryParam(t *testing.T) {
	if testHandler == nil {
		t.Skip("engine not initialised; set DATA_FILE")
	}

	r, err := http.NewRequest("GET", "/?client-ip=8.8.8.8", nil)
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}

	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, r)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected HTTP 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "8.8.8.8") {
		t.Error("Expected page to contain the queried IP address 8.8.8.8")
	}
}
