import yaml

with open('/workspaces/taskmaster/srcs/config.yml', 'r') as file:
    prime_service = yaml.safe_load(file)

print(prime_service['programs']['nginx']['cmd'])
print(prime_service['programs']['vogsphere']['cmd'])