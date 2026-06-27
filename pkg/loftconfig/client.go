package loftconfig

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/loft-sh/devpod/pkg/credentials"
	"github.com/loft-sh/devpod/pkg/log"
	"github.com/loft-sh/devpod/pkg/platform/client"
)

func GetLoftConfig(context, provider string, port int, logger log.Logger) (*client.Config, error) {
	request := &LoftConfigRequest{
		Context:  context,
		Provider: provider,
	}

	rawJson, err := json.Marshal(request)
	if err != nil {
		logger.Errorf("Error parsing request: %v", err)
		return nil, err
	}

	configResponse := &LoftConfigResponse{}
	out, err := credentials.PostWithRetry(port, "loft-platform-credentials", bytes.NewReader(rawJson), logger)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(out, configResponse)
	if err != nil {
		return nil, fmt.Errorf("decode loft config %s: %w", string(out), err)
	}

	return configResponse.LoftConfig, nil
}
