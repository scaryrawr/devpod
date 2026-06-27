package config

import "encoding/json"

type ImageMetadataConfig struct {
	Raw    []*ImageMetadata
	Config []*ImageMetadata
}

type ImageMetadata struct {
	ID                     string `json:"id,omitempty"`
	Entrypoint             string `json:"entrypoint,omitempty"`
	DevContainerConfigBase `json:",inline"`
	DevContainerActions    `json:",inline"`
	NonComposeBase         `json:",inline"`
}

func (m *ImageMetadata) UnmarshalJSON(data []byte) error {
	type imageMetadata ImageMetadata
	var raw imageMetadata
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*m = ImageMetadata(raw)
	return applyLegacyPortsAttributes(data, &m.DevContainerConfigBase)
}
