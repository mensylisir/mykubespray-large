find ./ -type f -name "*.yaml" -exec sed -i '/- :v0.8.6/d' {} +
