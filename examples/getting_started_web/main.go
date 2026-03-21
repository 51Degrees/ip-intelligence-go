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

// Package main demonstrates using 51Degrees IP Intelligence in an HTTP handler
// with a full HTML UI including an interactive map.
//
// The server listens on port 8080 and renders an HTML page showing IP
// intelligence results with a Leaflet map displaying the detected location.
// Users can look up arbitrary IP addresses via a form or the "client-ip"
// query parameter.
//
// Required environment variables:
//
//	DATA_FILE    Path to the IP Intelligence .ipi data file
//
// Run:
//
//	DATA_FILE=./51Degrees-EnterpriseIpiV41.ipi go run ./examples/getting_started_web
//
// Then open http://localhost:8080/ in a browser, or:
//
//	curl "http://localhost:8080/?client-ip=185.28.167.77"
package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/51Degrees/ip-intelligence-go/v4/ipi_interop"
	"github.com/51Degrees/ip-intelligence-go/v4/ipi_onpremise"
)

// ipiHandler holds a reference to the IPI engine and implements http.Handler.
type ipiHandler struct {
	engine *ipi_onpremise.Engine
}

// ServeHTTP extracts the client IP (consulting query parameters first), queries
// the IPI engine, and renders the HTML results page.
func (h *ipiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	queryParams := extractQueryParams(r)
	clientIP := extractClientIP(r, queryParams)

	values, err := h.engine.Process(clientIP)
	if err != nil {
		log.Printf("IP Intelligence error: %v", err)
		http.Error(w, "IP Intelligence processing error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := fmt.Fprint(w, renderHTML(clientIP, values)); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

// extractQueryParams returns URL query parameters as a simple map (first value wins).
func extractQueryParams(r *http.Request) map[string]string {
	q := r.URL.Query()
	params := make(map[string]string, len(q))
	for key, vals := range q {
		if len(vals) > 0 {
			params[key] = vals[0]
		}
	}
	return params
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

// prop is a helper that returns the string value of a property, or "Unknown".
func prop(values ipi_interop.Values, name string) string {
	if values == nil {
		return "Unknown"
	}
	val, found := values.GetValueByProperty(name)
	if !found || val == nil {
		return "Unknown"
	}
	return fmt.Sprintf("%v", val)
}

// propWeighted returns a weighted property value formatted with its weight.
func propWeighted(values ipi_interop.Values, name string) string {
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

// renderHTML builds the full HTML page from the IPI results.
func renderHTML(clientIP string, values ipi_interop.Values) string {
	get := func(name string) string { return prop(values, name) }

	registeredName := get("RegisteredName")
	registeredOwner := get("RegisteredOwner")
	registeredCountry := get("RegisteredCountry")
	ipRangeStart := get("IpRangeStart")
	ipRangeEnd := get("IpRangeEnd")
	country := get("Country")
	countryCode := get("CountryCode")
	countryCode3 := get("CountryCode3")
	region := get("Region")
	state := get("State")
	town := get("Town")
	latitude := get("Latitude")
	longitude := get("Longitude")
	areas := get("Areas")
	accuracyRadius := get("AccuracyRadiusMin")
	timeZoneOffset := get("TimeZoneOffset")
	mcc := propWeighted(values, "Mcc")

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>IP Intelligence Web Example</title>
    <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" />
    <style>
        body { margin: 2em; font-family: sans-serif; }
        table { width: 100%%; max-width: 800px; border-collapse: collapse; font-size: smaller; }
        th, td { text-align: left; border-bottom: 1px solid #ddd; padding: 5px; }
        .lightyellow { background-color: lightyellow; }
        .lightgreen { background-color: lightgreen; }
        .smaller { font-size: smaller; }
        details { margin-top: 20px; }
        summary { cursor: pointer; padding: 10px; background-color: #f5f5f5; border-radius: 5px; }
        summary:hover { background-color: #e8e8e8; }
    </style>
</head>
<body>
    <h2>IP Intelligence Web Example</h2>
    <p>This example demonstrates 51Degrees on-premise IP Intelligence lookups in a Go web application.</p>

    <form method="get" action="" style="margin-bottom: 25px;">
        <label for="client-ip" style="font-weight: bold;">IP Address Lookup:</label>
        <input type="text" id="client-ip" name="client-ip" value="%s"
               placeholder="e.g., 8.8.8.8" style="margin-left: 10px; padding: 8px; width: 200px; border: 1px solid #ccc; border-radius: 4px;">
        <button type="submit" style="margin-left: 10px; padding: 8px 20px; background-color: #007cba; color: white; border: none; border-radius: 4px; cursor: pointer;">Look Up</button>
    </form>

    <h3>IP Intelligence Results</h3>
    <table>
        <tr><th>Property</th><th>Value</th></tr>
        <tr class="lightyellow"><td><b>Registered Name</b></td><td>%s</td></tr>
        <tr class="lightgreen"><td><b>Registered Owner</b></td><td>%s</td></tr>
        <tr class="lightyellow"><td><b>Registered Country</b></td><td>%s</td></tr>
        <tr class="lightgreen"><td><b>IP Range Start</b></td><td>%s</td></tr>
        <tr class="lightyellow"><td><b>IP Range End</b></td><td>%s</td></tr>
        <tr class="lightgreen"><td><b>Country</b></td><td>%s</td></tr>
        <tr class="lightyellow"><td><b>Country Code</b></td><td>%s</td></tr>
        <tr class="lightgreen"><td><b>Country Code 3</b></td><td>%s</td></tr>
        <tr class="lightyellow"><td><b>Region</b></td><td>%s</td></tr>
        <tr class="lightgreen"><td><b>State</b></td><td>%s</td></tr>
        <tr class="lightyellow"><td><b>Town</b></td><td>%s</td></tr>
        <tr class="lightgreen"><td><b>Latitude</b></td><td>%s</td></tr>
        <tr class="lightyellow"><td><b>Longitude</b></td><td>%s</td></tr>
        <tr class="lightgreen"><td><b>Areas</b></td><td>%s</td></tr>
        <tr class="lightyellow"><td><b>Accuracy Radius</b></td><td>%s</td></tr>
        <tr class="lightgreen"><td><b>Time Zone Offset</b></td><td>%s</td></tr>
        <tr class="lightyellow"><td><b>MCC</b></td><td>%s</td></tr>
    </table>

    <br/>
    <div id="map-section" style="display: none;">
        <h3>Location Map</h3>
        <div id="map" style="height: 400px; width: 100%%; max-width: 800px; border: 1px solid #ccc;"></div>
    </div>

    <script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>
    <script src="https://unpkg.com/wellknown@0.5.0/wellknown.js"></script>
    <script>
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
    </script>
</body>
</html>`,
		escapeHTML(clientIP),
		escapeHTML(registeredName),
		escapeHTML(registeredOwner),
		escapeHTML(registeredCountry),
		escapeHTML(ipRangeStart),
		escapeHTML(ipRangeEnd),
		escapeHTML(country),
		escapeHTML(countryCode),
		escapeHTML(countryCode3),
		escapeHTML(region),
		escapeHTML(state),
		escapeHTML(town),
		escapeHTML(latitude),
		escapeHTML(longitude),
		escapeHTML(areas),
		escapeHTML(accuracyRadius),
		escapeHTML(timeZoneOffset),
		escapeHTML(mcc),
		// JS values
		escapeJS(areas),
		escapeJS(latitude),
		escapeJS(longitude),
	)
}

func main() {
	dataFile := os.Getenv("DATA_FILE")
	if dataFile == "" {
		dataFile = "51Degrees-EnterpriseIpiV41.ipi"
	}

	config := ipi_interop.NewConfigIpi(ipi_interop.LowMemory)
	engine, err := ipi_onpremise.New(
		ipi_onpremise.WithConfigIpi(config),
		ipi_onpremise.WithDataFile(dataFile),
		ipi_onpremise.WithAutoUpdate(false),
		ipi_onpremise.WithTempDataCopy(false),
	)
	if err != nil {
		log.Fatalf("Failed to create IP Intelligence engine: %v", err)
	}
	defer engine.Stop()

	h := &ipiHandler{engine: engine}
	http.Handle("/", h)

	const addr = ":8080"
	log.Printf("IP Intelligence Web Example — listening on %s", addr)
	log.Printf("Data file: %s", dataFile)
	log.Fatal(http.ListenAndServe(addr, nil))
}
