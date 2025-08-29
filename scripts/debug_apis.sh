#!/bin/bash

echo "=== NOAA K-Index API ==="
curl -s "https://services.swpc.noaa.gov/products/noaa-planetary-k-index.json" | head -5

echo -e "\n=== NOAA Solar API ==="
curl -s "https://services.swpc.noaa.gov/products/solar-wind/plasma-7-day.json" | head -5

echo -e "\n=== N0NBH API ==="
curl -s "https://www.hamqsl.com/solarxml.php" | head -5

echo -e "\n=== SIDC RSS ==="
curl -s "https://www.sidc.be/silso/INFO/snmtotcsv.php" | head -5
