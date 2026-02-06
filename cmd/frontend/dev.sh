#!/bin/bash

echo "Step 1: Clear Libs..."
rm -rf ../backend/static/libs/*

echo "Step 2: Copying LoadJS..."
cp -r node_modules/loadjs/dist/ ../backend/static/libs/loadjs/

echo "Step 3: Copying jQuery..."
cp -r node_modules/jquery/dist/ ../backend/static/libs/jquery/

echo "Step 4: Copying Clipboard..."  
cp -r node_modules/clipboard/dist/ ../backend/static/libs/clipboard/

echo "Step 5: Copying Select2..."
cp -r node_modules/select2/dist/ ../backend/static/libs/select2/

echo "Step 6: Copying Select2 Theme..."
cp -r node_modules/select2-tailwindcss-v4-theme/dist/ ../backend/static/libs/select2-tailwindcss-v4-theme/

echo "Step 7: Copying Echarts..."
cp -r node_modules/echarts/dist/ ../backend/static/libs/echarts/

echo "Step 8: Copying Dayjs..."
mkdir -p ../backend/static/libs/dayjs/ 
cp node_modules/dayjs/dayjs.min.js ../backend/static/libs/dayjs/
cp node_modules/dayjs/plugin/relativeTime.js ../backend/static/libs/dayjs/relativeTime.js
cp node_modules/dayjs/locale/zh-cn.js ../backend/static/libs/dayjs/zh-cn.js

echo "Step 9: Running TailwindCSS..."
pnpm tailwindcss -i ./css/input.css -o ../backend/static/css/style.css --watch --poll --verbose
