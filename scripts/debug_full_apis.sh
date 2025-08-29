#!/bin/bash

echo "=== NOAA K-Index API (Full) ==="
curl -s "https://services.swpc.noaa.gov/products/noaa-planetary-k-index.json" > /tmp/noaa_k.json
echo "Response saved to /tmp/noaa_k.json"
head -20 /tmp/noaa_k.json

echo -e "\n=== NOAA Solar API (Full) ==="
curl -s "https://services.swpc.noaa.gov/products/solar-wind/plasma-7-day.json" > /tmp/noaa_solar.json
echo "Response saved to /tmp/noaa_solar.json"
head -20 /tmp/noaa_solar.json

echo -e "\n=== N0NBH API (Full XML) ==="
curl -s "https://www.hamqsl.com/solarxml.php" > /tmp/n0nbh.xml
echo "Response saved to /tmp/n0nbh.xml"
head -20 /tmp/n0nbh.xml

echo -e "\n=== SIDC RSS (Check redirect) ==="
curl -I "https://www.sidc.be/silso/INFO/snmtotcsv.php" 2>&1 | head -10
