package utils

import (
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"gopkg.in/yaml.v2"
)

type ApiServerClient struct {
	ApiServerConf entity.ApiServerOidcConf
	OSClient      OSClient
}

func NewApiServerClient(conf entity.ApiServerOidcConf, osClient OSClient) *ApiServerClient {
	return &ApiServerClient{
		ApiServerConf: conf,
		OSClient:      osClient,
	}
}

func (client *ApiServerClient) ModifyConfig() error {
	configFile := "/etc/kubernetes/manifests/kube-apiserver.yaml"
	err := client.OSClient.Chmod(configFile, "0644")
	if err != nil {
		logger.GetLogger().Errorf("Chmod %s failed", configFile)
		return err
	}
	data, err := client.OSClient.ReadBytes(configFile)
	if err != nil {
		logger.GetLogger().Errorf("Read %s failed", configFile)
		return err
	}
	var configData map[string]interface{}
	err = yaml.Unmarshal(data, &configData)
	if err != nil {
		logger.GetLogger().Errorf("unmarshal data %v failed", string(data))
		return err
	}
	updateCommandInContainer(configData, "kube-apiserver", []string{
		"--oidc-issuer-url=" + client.ApiServerConf.OIDCIssuerUrl,
		"--oidc-client-id=" + client.ApiServerConf.OIDCClientID,
		"--oidc-username-claim=name",
		"--oidc-username-prefix=-",
		"--oidc-groups-claim=groups",
		"--oidc-ca-file=" + client.ApiServerConf.OIDCCAFile,
	})

	updateData, err := yaml.Marshal(&configData)
	if err != nil {
		logger.GetLogger().Errorf("Marshal %v failed", configData)
		return err
	}

	err = client.OSClient.WriteFile(string(updateData), configFile)
	if err != nil {
		logger.GetLogger().Errorf("Write %v failed", configFile)
		return err
	}
	err = client.OSClient.Chmod(configFile, "0644")
	if err != nil {
		logger.GetLogger().Errorf("Chmod %v failed", configFile)
		return err
	}
	return nil
}

func updateCommandInContainer(configData map[string]interface{}, containerName string, newCommands []string) {
	spec, ok := configData["spec"].(map[interface{}]interface{})
	if ok {
		if containers, ok := spec["containers"].([]interface{}); ok {
			for _, containerRaw := range containers {
				if container, ok := containerRaw.(map[interface{}]interface{}); ok {
					if name, ok := container["name"].(string); ok && name == containerName {
						tempCommands := container["command"].([]interface{})
						if command, ok := container["command"].([]interface{}); ok {
							for _, newCommand := range newCommands {
								if !contains(command, newCommand) {
									tempCommands = append(tempCommands, newCommand)
								}
							}
						} else {
							// 如果没有 command 字段，则初始化
							container["command"] = toInterfaceSlice(newCommands)
						}
						container["command"] = tempCommands
					}
				}
			}
		}
	}
}
