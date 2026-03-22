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

// Package main demonstrates using 51Degrees Device Detection and IP Intelligence
// engines together inside an HTTP handler with a full HTML UI including an
// interactive map.
//
// The server listens on port 8080 and renders an HTML page showing:
//   - Device detection results (hardware, platform, browser) in a left column
//   - IP intelligence results (location, network) in a right column
//   - An interactive Leaflet map showing the detected location
//   - An IP address lookup form for custom lookups
//
// Both engines are queried concurrently for each request.
//
// Required environment variables:
//
//	DATA_FILE    Path to the IP Intelligence .ipi data file
//	DD_DATA_FILE Path to the Device Detection .hash data file
//
// Run:
//
//	DATA_FILE=./51Degrees-EnterpriseIpiV41.ipi \
//	DD_DATA_FILE=./51Degrees-LiteV4.1.hash \
//	go run ./examples/mixed/getting_started_web_ui
//
// Then open http://localhost:8080/ in a browser.
package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/51Degrees/device-detection-go/v4/dd"
	ddOnpremise "github.com/51Degrees/device-detection-go/v4/onpremise"
	"github.com/51Degrees/ip-intelligence-go/v4/ipi_interop"
	"github.com/51Degrees/ip-intelligence-go/v4/ipi_onpremise"
)

// propertyDef describes a single property row in a results table.
type propertyDef struct {
	label    string // display label
	name     string // engine property name
	weighted bool   // true for properties that carry a weight (e.g. Mcc)
}

// ddDisplayProperties defines the DD properties shown in the results table.
var ddDisplayProperties = []propertyDef{
	{label: "Hardware Name", name: "HardwareName"},
	{label: "Hardware Model", name: "HardwareModel"},
	{label: "Platform Name", name: "PlatformName"},
	{label: "Platform Version", name: "PlatformVersion"},
	{label: "Browser Name", name: "BrowserName"},
	{label: "Browser Version", name: "BrowserVersion"},
}

// ipiDisplayProperties defines the IPI properties shown in the results table.
var ipiDisplayProperties = []propertyDef{
	{label: "Registered Name", name: "RegisteredName"},
	{label: "Registered Owner", name: "RegisteredOwner"},
	{label: "Registered Country", name: "RegisteredCountry"},
	{label: "IP Range Start", name: "IpRangeStart"},
	{label: "IP Range End", name: "IpRangeEnd"},
	{label: "Country", name: "Country"},
	{label: "Country Code", name: "CountryCode"},
	{label: "Country Code 3", name: "CountryCode3"},
	{label: "Region", name: "Region"},
	{label: "State", name: "State"},
	{label: "Town", name: "Town"},
	{label: "Latitude", name: "Latitude"},
	{label: "Longitude", name: "Longitude"},
	{label: "Areas", name: "Areas"},
	{label: "Accuracy Radius", name: "AccuracyRadiusMin"},
	{label: "Time Zone Offset", name: "TimeZoneOffset"},
	{label: "MCC", name: "Mcc", weighted: true},
}

// ddResponseProperties lists DD property names passed to the engine (derived from ddDisplayProperties).
var ddResponseProperties = func() []string {
	names := make([]string, len(ddDisplayProperties))
	for i, p := range ddDisplayProperties {
		names[i] = p.name
	}
	return names
}()

// mixedHandler holds references to both engines and implements http.Handler.
type mixedHandler struct {
	ddEngine  *ddOnpremise.Engine
	ipiEngine *ipi_onpremise.Engine
}

// ServeHTTP extracts evidence from the request, queries both engines concurrently,
// and renders an HTML page with the results.
func (h *mixedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	queryParams := extractQueryParams(r)
	evidence := extractEvidence(r, queryParams)
	clientIP := extractClientIP(r, queryParams)

	var (
		ddResults     *dd.ResultsHash
		ipiValues     ipi_interop.Values
		ddErr, ipiErr error
		wg            sync.WaitGroup
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		ddResults, ddErr = h.ddEngine.Process(evidence)
	}()
	go func() {
		defer wg.Done()
		ipiValues, ipiErr = h.ipiEngine.Process(clientIP)
	}()
	wg.Wait()

	if ddErr != nil {
		log.Printf("Device Detection error: %v", ddErr)
	} else {
		defer ddResults.Free()
	}
	if ipiErr != nil {
		log.Printf("IP Intelligence error: %v", ipiErr)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := fmt.Fprint(w, renderHTML(clientIP, ddResults, ipiValues, r)); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

// extractQueryParams returns URL query parameters as a simple map (first value wins).
func extractQueryParams(r *http.Request) map[string]string {
	q := r.URL.Query()
	params := make(map[string]string, len(q))
	for key, values := range q {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}
	return params
}

// extractEvidence converts HTTP request headers and URL query parameters
// into a DD evidence slice.
func extractEvidence(r *http.Request, queryParams map[string]string) []ddOnpremise.Evidence {
	evidence := make([]ddOnpremise.Evidence, 0, len(r.Header)+len(queryParams))
	for key, values := range r.Header {
		if len(values) > 0 {
			evidence = append(evidence, ddOnpremise.Evidence{
				Prefix: dd.HttpHeaderString,
				Key:    key,
				Value:  values[0],
			})
		}
	}
	for key, value := range queryParams {
		evidence = append(evidence, ddOnpremise.Evidence{
			Prefix: dd.HttpEvidenceQuery,
			Key:    key,
			Value:  value,
		})
	}
	return evidence
}

// extractClientIP determines the client IP from the request, consulting
// query parameters, X-Forwarded-For and X-Real-Ip headers before falling
// back to RemoteAddr.
func extractClientIP(r *http.Request, queryParams map[string]string) string {
	if clientIP, ok := queryParams["client-ip"]; ok && clientIP != "" {
		return strings.TrimSpace(clientIP)
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])
	}
	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// ddProp returns the string value of a device detection property, or "Unknown".
func ddProp(results *dd.ResultsHash, name string) string {
	if results == nil {
		return "Unknown"
	}
	hasValues, err := results.HasValues(name)
	if err != nil || !hasValues {
		return "Unknown"
	}
	value, err := results.ValuesString(name, ",")
	if err != nil {
		return "Unknown"
	}
	return value
}

// ipiProp returns the string value of an IPI property, or "Unknown".
func ipiProp(values ipi_interop.Values, name string) string {
	if values == nil {
		return "Unknown"
	}
	val, found := values.GetValueByProperty(name)
	if !found || val == nil {
		return "Unknown"
	}
	return fmt.Sprintf("%v", val)
}

// ipiPropWeighted returns a weighted property value formatted with its weight.
func ipiPropWeighted(values ipi_interop.Values, name string) string {
	if values == nil {
		return "Unknown"
	}
	val, weight, found := values.GetValueWeightByProperty(name)
	if !found || val == nil {
		return "Unknown"
	}
	return fmt.Sprintf("%v (weight: %.2f)", val, weight)
}

// escapeHTML performs minimal escaping for safe HTML embedding.
func escapeHTML(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&#39;")
	return r.Replace(s)
}

// escapeJS escapes a string for safe embedding in a JavaScript single-quoted string literal.
func escapeJS(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `'`, `\'`, `"`, `\"`, "\n", `\n`, "\r", `\r`)
	return r.Replace(s)
}

// buildTableRows generates HTML table rows from a property list using the provided getter.
func buildTableRows(props []propertyDef, getter func(propertyDef) string) string {
	colors := [2]string{"lightyellow", "lightgreen"}
	var sb strings.Builder
	for i, p := range props {
		value := escapeHTML(getter(p))
		fmt.Fprintf(&sb, "                <tr class=\"%s\"><td><b>%s</b></td><td>%s</td></tr>\n",
			colors[i%2], escapeHTML(p.label), value)
	}
	return sb.String()
}

// buildEvidenceRows builds HTML table rows for the evidence section.
func buildEvidenceRows(r *http.Request) string {
	var sb strings.Builder
	i := 0
	for key, values := range r.Header {
		if len(values) > 0 {
			cls := "lightyellow"
			if i%2 == 1 {
				cls = "lightgreen"
			}
			sb.WriteString(fmt.Sprintf(
				`<tr class="%s"><td>%s</td><td>%s</td></tr>`,
				cls, escapeHTML(key), escapeHTML(values[0])))
			i++
		}
	}
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			cls := "lightyellow"
			if i%2 == 1 {
				cls = "lightgreen"
			}
			sb.WriteString(fmt.Sprintf(
				`<tr class="%s"><td><b>query.%s</b></td><td>%s</td></tr>`,
				cls, escapeHTML(key), escapeHTML(values[0])))
			i++
		}
	}
	return sb.String()
}

// renderHTML builds the full HTML page from both engines' results.
func renderHTML(clientIP string, ddResults *dd.ResultsHash, ipiValues ipi_interop.Values, r *http.Request) string {
	ddRows := buildTableRows(ddDisplayProperties, func(p propertyDef) string {
		return ddProp(ddResults, p.name)
	})
	ipiRows := buildTableRows(ipiDisplayProperties, func(p propertyDef) string {
		if p.weighted {
			return ipiPropWeighted(ipiValues, p.name)
		}
		return ipiProp(ipiValues, p.name)
	})

	latitude := ipiProp(ipiValues, "Latitude")
	longitude := ipiProp(ipiValues, "Longitude")
	areas := ipiProp(ipiValues, "Areas")
	evidenceRows := buildEvidenceRows(r)

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Device Detection + IP Intelligence Example</title>
    <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" />
    <style>
        body { margin: 2em; font-family: sans-serif; }
        table { width: 100%%; border-collapse: collapse; font-size: smaller; }
        th, td { text-align: left; border-bottom: 1px solid #ddd; padding: 5px; }
        .lightyellow { background-color: lightyellow; }
        .lightgreen { background-color: lightgreen; }
        .smaller { font-size: smaller; }
        .results-columns { display: flex; gap: 20px; }
        .results-columns > div { flex: 1; }
        details { margin-top: 20px; }
        summary { cursor: pointer; padding: 10px; background-color: #f5f5f5; border-radius: 5px; }
        summary:hover { background-color: #e8e8e8; }
        @media (max-width: 768px) {
            body { margin: 1em; }
            .results-columns { flex-direction: column; gap: 0; }
        }
    </style>
</head>
<body>
    <h2>Combined Device Detection and IP Intelligence Example</h2>

    <p>
        This example demonstrates the use of 51Degrees Device Detection and IP Intelligence
        engines together in a single Go web application. It highlights:
    </p>
    <ol>
        <li>Device detection using User-Agent and request headers</li>
        <li>IP-based geolocation and network information</li>
        <li>Parallel processing of both engines for optimal performance</li>
    </ol>

    <form method="get" action="" style="margin-bottom: 25px;">
        <label for="client-ip" style="font-weight: bold;">IP Address Lookup:</label>
        <input type="text" id="client-ip" name="client-ip" value="%s"
               placeholder="e.g., 8.8.8.8" style="margin-left: 10px; padding: 8px; width: 200px; border: 1px solid #ccc; border-radius: 4px;">
        <button type="submit" style="margin-left: 10px; padding: 8px 20px; background-color: #007cba; color: white; border: none; border-radius: 4px; cursor: pointer;">Look Up</button>
    </form>

    <div class="results-columns">
        <div>
            <h3>Device Detection Results</h3>
            <table>
                <tr><th>Property</th><th>Value</th></tr>
%s            </table>
        </div>

        <div>
            <h3>IP Intelligence Results</h3>
            <table>
                <tr><th>Property</th><th>Value</th></tr>
%s            </table>
        </div>
    </div>

    <br/>

    <div id="map-section" style="display: none;">
        <h3>Location Map</h3>
        <div id="map" style="height: 400px; width: 100%%; border: 1px solid #ccc;"></div>
    </div>

    <br/>

    <div id="evidence">
        <h3>Evidence Used</h3>
        <p class="smaller">Evidence collected from the current request</p>
        <details>
            <summary>Click to view evidence details</summary>
            <table style="margin-top: 10px;">
                <tr><th>Key</th><th>Value</th></tr>
                %s
            </table>
        </details>
    </div>

    <script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>
    <script src="https://unpkg.com/wellknown@0.5.0/wellknown.js"></script>
    <script>
        (function() {
            var UNKNOWN = 'Unknown';
            var EMPTY_POLYGON = 'POLYGON EMPTY';

            function isValid(v) {
                return v && v !== UNKNOWN && v !== '0' && v !== '';
            }

            var areasWkt = '%s';
            var latitude = '%s';
            var longitude = '%s';

            if (isValid(areasWkt) && areasWkt !== EMPTY_POLYGON) {
                document.getElementById('map-section').style.display = 'block';
                var map = L.map('map');
                L.tileLayer('https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png', {
                    attribution: '&copy; OpenStreetMap contributors &copy; CARTO',
                    subdomains: 'abcd',
                    maxZoom: 19
                }).addTo(map);
                try {
                    var geoJson = wellknown.parse(areasWkt);
                    if (geoJson) {
                        var polygon = L.geoJSON(geoJson, {
                            style: { color: '#ff0000', weight: 2, opacity: 0.8, fillColor: '#ff0000', fillOpacity: 0.2 }
                        }).addTo(map);
                        map.fitBounds(polygon.getBounds());
                        if (isValid(latitude) && isValid(longitude)) {
                            var lat = parseFloat(latitude);
                            var lng = parseFloat(longitude);
                            if (!isNaN(lat) && !isNaN(lng)) {
                                L.marker([lat, lng]).addTo(map)
                                    .bindPopup('IP Location<br>Lat: ' + lat + '<br>Lng: ' + lng)
                                    .openPopup();
                            }
                        }
                    }
                } catch (e) {
                    console.error('Error parsing polygon:', e);
                    document.getElementById('map-section').style.display = 'none';
                }
            } else if (isValid(latitude) && isValid(longitude)) {
                document.getElementById('map-section').style.display = 'block';
                var lat = parseFloat(latitude);
                var lng = parseFloat(longitude);
                if (!isNaN(lat) && !isNaN(lng)) {
                    var map = L.map('map').setView([lat, lng], 10);
                    L.tileLayer('https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png', {
                        attribution: '&copy; OpenStreetMap contributors &copy; CARTO',
                        subdomains: 'abcd',
                        maxZoom: 19
                    }).addTo(map);
                    L.marker([lat, lng]).addTo(map)
                        .bindPopup('IP Location<br>Lat: ' + lat + '<br>Lng: ' + lng)
                        .openPopup();
                }
            }
        })();
    </script>
</body>
</html>`,
		escapeHTML(clientIP),
		ddRows,
		ipiRows,
		evidenceRows,
		// JS values for map
		escapeJS(areas),
		escapeJS(latitude),
		escapeJS(longitude),
	)
}

func main() {
	ipiDataFile := os.Getenv("DATA_FILE")
	ddDataFile := os.Getenv("DD_DATA_FILE")

	ddConfig := dd.NewConfigHash(dd.LowMemory)
	ddEngine, err := ddOnpremise.New(
		ddOnpremise.WithConfigHash(ddConfig),
		ddOnpremise.WithDataFile(ddDataFile),
		ddOnpremise.WithAutoUpdate(false),
		ddOnpremise.WithTempDataCopy(false),
		ddOnpremise.WithProperties(ddResponseProperties),
	)
	if err != nil {
		log.Fatalf("Failed to create Device Detection engine: %v", err)
	}
	defer ddEngine.Stop()

	ipiConfig := ipi_interop.NewConfigIpi(ipi_interop.LowMemory)
	ipiEngine, err := ipi_onpremise.New(
		ipi_onpremise.WithConfigIpi(ipiConfig),
		ipi_onpremise.WithDataFile(ipiDataFile),
		ipi_onpremise.WithAutoUpdate(false),
		ipi_onpremise.WithTempDataCopy(false),
	)
	if err != nil {
		log.Fatalf("Failed to create IP Intelligence engine: %v", err)
	}
	defer ipiEngine.Stop()

	h := &mixedHandler{ddEngine: ddEngine, ipiEngine: ipiEngine}
	http.Handle("/", h)

	const addr = ":8080"
	log.Printf("Mixed Getting Started Web UI Example — listening on %s", addr)
	log.Printf("Device Detection data: %s", ddDataFile)
	log.Printf("IP Intelligence data:  %s", ipiDataFile)
	log.Fatal(http.ListenAndServe(addr, nil))
}
