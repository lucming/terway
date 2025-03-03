package daemon

import (
	"os"

	"github.com/AliyunContainerService/terway/types"
	"github.com/AliyunContainerService/terway/types/route"
	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
)

// Config configuration of terway daemon
type Config struct {
	Version                string              `yaml:"version" json:"version"`
	AccessID               string              `yaml:"access_key" json:"access_key"`
	AccessSecret           string              `yaml:"access_secret" json:"access_secret"`
	CredentialPath         string              `yaml:"credential_path" json:"credential_path"`
	ServiceCIDR            string              `yaml:"service_cidr" json:"service_cidr"`
	VSwitches              map[string][]string `yaml:"vswitches" json:"vswitches"`
	ENITags                map[string]string   `yaml:"eni_tags" json:"eni_tags"`
	MaxPoolSize            int                 `yaml:"max_pool_size" json:"max_pool_size"`
	MinPoolSize            int                 `yaml:"min_pool_size" json:"min_pool_size"`
	MinENI                 int                 `yaml:"min_eni" json:"min_eni"`
	MaxENI                 int                 `yaml:"max_eni" json:"max_eni"`
	Prefix                 string              `yaml:"prefix" json:"prefix"`
	SecurityGroup          string              `yaml:"security_group" json:"security_group"`
	SecurityGroups         []string            `yaml:"security_groups" json:"security_groups"`
	EniCapRatio            float64             `yaml:"eni_cap_ratio" json:"eni_cap_ratio"`
	EniCapShift            int                 `yaml:"eni_cap_shift" json:"eni_cap_shift"`
	VSwitchSelectionPolicy string              `yaml:"vswitch_selection_policy" json:"vswitch_selection_policy"`
	EnableEIPPool          string              `yaml:"enable_eip_pool" json:"enable_eip_pool"`
	IPStack                string              `yaml:"ip_stack" json:"ip_stack" validate:"oneof=ipv4 ipv6 dual" mod:"default=ipv4"` // default ipv4 , support ipv4 dual
	// rob the eip instance even the eip already bound to other resource
	AllowEIPRob                 string                  `yaml:"allow_eip_rob" json:"allow_eip_rob"`
	EnableENITrunking           bool                    `yaml:"enable_eni_trunking" json:"enable_eni_trunking"`
	CustomStatefulWorkloadKinds []string                `yaml:"custom_stateful_workload_kinds" json:"custom_stateful_workload_kinds"`
	IPAMType                    types.IPAMType          `yaml:"ipam_type" json:"ipam_type"`           // crd or default
	ENICapPolicy                types.ENICapPolicy      `yaml:"eni_cap_policy" json:"eni_cap_policy"` // prefer trunk or secondary
	BackoffOverride             map[string]wait.Backoff `json:"backoff_override,omitempty"`
	ExtraRoutes                 []route.Route           `json:"extra_routes,omitempty"`
	DisableDevicePlugin         bool                    `json:"disable_device_plugin"`
	WaitTrunkENI                bool                    `json:"wait_trunk_eni"` // true for don't create trunk eni
	ENITagFilter                map[string]string       `json:"eni_tag_filter"` // if set , only enis match filter, will be managed
	DisableSecurityGroupCheck   bool                    `json:"disable_security_group_check"`
	KubeClientQPS               float32                 `json:"kube_client_qps"`
	KubeClientBurst             int                     `json:"kube_client_burst"`
}

func (c *Config) GetSecurityGroups() []string {
	sgIDs := sets.NewString()
	if c.SecurityGroup != "" {
		sgIDs.Insert(c.SecurityGroup)
	}
	sgIDs.Insert(c.SecurityGroups...)
	return sgIDs.List()
}

func (c *Config) GetVSwitchIDs() []string {
	var vsws []string
	for _, ids := range c.VSwitches {
		vsws = append(vsws, ids...)
	}
	return vsws
}

func (c *Config) GetExtraRoutes() []string {
	var vsws []string
	for _, ids := range c.VSwitches {
		vsws = append(vsws, ids...)
	}
	return vsws
}

// GetConfigFromFileWithMerge parse Config from file
func GetConfigFromFileWithMerge(filePath string, cfg []byte) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return MergeConfigAndUnmarshal(cfg, data)
}

func MergeConfigAndUnmarshal(topCfg, baseCfg []byte) (*Config, error) {
	if len(topCfg) == 0 { // no topCfg, unmarshal baseCfg and return
		config := &Config{}
		err := json.Unmarshal(baseCfg, config)
		return config, err
	}

	// MergePatch in RFC7396
	jsonBytes, err := jsonpatch.MergePatch(baseCfg, topCfg)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = json.Unmarshal(jsonBytes, config)

	return config, err
}
