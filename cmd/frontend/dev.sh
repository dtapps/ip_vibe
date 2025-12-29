#!/bin/bash

echo "Step 1: Copying jQuery..."
cp -r node_modules/jquery/dist/ ../backend/static/libs/jquery/

echo "Step 2: Copying Clipboard..."  
cp -r node_modules/clipboard/dist/ ../backend/static/libs/clipboard/

echo "Step 3: Copying Select2..."
cp -r node_modules/select2/dist/ ../backend/static/libs/select2/

# echo "Step 4: Copying Select2 Theme..."
# cp -r node_modules/select2-tailwindcss-v4-theme/dist/ ../backend/static/libs/select2-tailwindcss-v4-theme/

echo "Step 5: Running TailwindCSS..."
pnpm tailwindcss -i ./css/input.css -o ../backend/static/css/style.css --watch --poll --verbose
