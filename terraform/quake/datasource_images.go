// Copyright (c) 2016-2020 Hewlett Packard Enterprise Development LP.

package quake

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceImage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceImageRead,
		Schema: map[string]*schema.Schema{
			avImages: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     imageResource(),
			},
			dsFilter: dataSourceFiltersSchema(),
		},
	}

}

func dataSourceImageRead(d *schema.ResourceData, meta interface{}) error {
	p := meta.(*Config)
	available := p.availableResources

	var images = make([]map[string]interface{}, 0, len(available.Images))
	for _, image := range available.Images {
		filters, err := getFilters(d)
		if err != nil {
			return err
		}
		matched := (len(filters) == 0)
		flavorMatch, categoryMatch, versionMatch := true, true, true
		for _, f := range filters {
			if f.name == iFlavor && !f.match(iFlavor, image.Flavor) {
				flavorMatch = false
			}
			if f.name == iCategory && !f.match(iCategory, image.Category) {
				categoryMatch = false
			}
			if f.name == iVersion && !f.match(iVersion, image.Version) {
				versionMatch = false
			}
		}
		if matched || (flavorMatch && categoryMatch && versionMatch) {
			iData := map[string]interface{}{
				iFlavor:   image.Flavor,
				iVersion:  image.Version,
				iCategory: image.Category,
				"id":      image.ID,
			}
			images = append(images, iData)
		}
	}
	if err := d.Set(avImages, images); err != nil {
		return err
	}
	d.SetId("images")
	return nil
}
