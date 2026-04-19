package opnsense

import (
	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	"github.com/EvilBit-Labs/opnDossier/pkg/schema/opnsense"
)

// SerializeVLANs maps CommonDevice VLAN entries onto opnsense.VLANs.
func SerializeVLANs(in []model.VLAN) opnsense.VLANs {
	out := opnsense.VLANs{}
	for _, v := range in {
		out.VLAN = append(out.VLAN, opnsense.VLAN{
			If:     v.PhysicalIf,
			Tag:    v.Tag,
			Descr:  v.Description,
			Vlanif: v.VLANIf,
		})
	}
	return out
}
