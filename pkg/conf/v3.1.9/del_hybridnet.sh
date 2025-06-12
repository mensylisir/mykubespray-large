find ./ -type f -name "*.yaml" -exec sed -i '/- :v0.8.6/d' {} +
find ./ -type f -name "*.yaml" -exec sed -i '/kata-deploy/d' {} +
find ./ -type f -name "*.yaml" -exec sed -i '/node-feature-discovery/d' {} +
