// (C) Copyright 2016-2021 Hewlett Packard Enterprise Development LP

package quake

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	rest "github.com/hpe-hcss/quake-client/v1/pkg/client"
)

const (
	pName    = "name"
	pProfile = "profile"
	pLimits  = "limits"

	pProjectName        = "project_name"
	pProjectDescription = "project_description"
	pCompany            = "company"
	pAddress            = "address"
	pEmail              = "email"
	pEmailVerified      = "email_verified"
	pPhoneNumber        = "phone_number"
	pPhoneVerified      = "phone_number_verified"

	pHosts           = "hosts"
	pVolumes         = "volumes"
	pVolumeCapacity  = "volume_capacity"
	pPrivateNetworks = "private_networks"
)

func limitsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		pHosts: {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Maximum number of host allowed in the team.",
		},
		pVolumes: {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Maximum number of volumes allowed in the team.",
		},
		pVolumeCapacity: {
			Type:        schema.TypeFloat,
			Optional:    true,
			Description: "Total allowable volume capacity (GiB) allowed in the team.",
		},
		pPrivateNetworks: {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Maximum number of private networks allowed in the team.",
		},
	}
}

func profileSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		pProjectName: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "A friendly name of the team.",
		},
		pProjectDescription: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "A friendly description of the team.",
		},
		pCompany: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The company associated with the team.",
		},
		pAddress: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The company address with the team.",
		},
		pEmail: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Email address.",
		},
		pEmailVerified: {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Email address has been validated.",
		},
		pPhoneNumber: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Phine number.",
		},
		pPhoneVerified: {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Phine number has been validated.",
		},
	}
}

func projectSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		pName: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "A friendly name of the project.",
		},

		pProfile: {
			// TODO the V2 SDK doesn't (yet) support TypeMap with Elem *Resource for nested objects
			// This is the currently recommended work-around. See
			// https://github.com/hashicorp/terraform-plugin-sdk/issues/155
			// https://github.com/hashicorp/terraform-plugin-sdk/issues/616
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "Team profile.",
			Elem: &schema.Resource{
				Schema: profileSchema(),
			},
		},
		pLimits: {
			// TODO the V2 SDK doesn't (yet) support TypeMap with Elem *Resource for nested objects
			// This is the currently recommended work-around. See
			// https://github.com/hashicorp/terraform-plugin-sdk/issues/155
			// https://github.com/hashicorp/terraform-plugin-sdk/issues/616
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "Resource limits applied to this team.",
			Elem: &schema.Resource{
				Schema: limitsSchema(),
			},
		},
	}
}

func ProjectResource() *schema.Resource {
	return &schema.Resource{
		Create: resourceQuattroProjectCreate,
		Read:   resourceQuattroProjectRead,
		Delete: resourceQuattroProjectDelete,
		Update: resourceQuattroProjectUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: projectSchema(),
	}
}

func resourceQuattroProjectCreate(d *schema.ResourceData, meta interface{}) (err error) {
	defer func() {
		var nErr = rest.GenericOpenAPIError{}
		if errors.As(err, &nErr) {
			err = fmt.Errorf("failed to create project %s: %w", strings.Trim(string(nErr.Body()), "\n "), err)

		}
	}()

	p, err := getConfigFromMeta(meta)
	if err != nil {
		return err
	}

	profile := d.Get(pProfile).(map[string]interface{})
	limits := d.Get(pLimits).(map[string]interface{})

	np := rest.NewProject{
		Name: d.Get(pName).(string),
	}
	if profile != nil {
		np.Profile = rest.Profile{
			Address:     safeString(profile[pAddress]),
			Company:     safeString(profile[pCompany]),
			Email:       safeString(profile[pEmail]),
			PhoneNumber: safeString(profile[pPhoneNumber]),
			TeamDesc:    safeString(profile[pProjectDescription]),
			TeamName:    safeString(profile[pProjectName]),
		}
	}

	if limits != nil {
		np.Limits = rest.Limits{
			Hosts:           int32(safeInt(limits[pHosts])),
			Volumes:         int32(safeInt(limits[pVolumes])),
			VolumeCapacity:  int64(safeFloat(limits[pVolumeCapacity])),
			PrivateNetworks: int32(safeInt(limits[pPrivateNetworks])),
		}
	}

	ctx := p.GetContext()
	project, _, err := p.Client.ProjectsApi.Add(ctx, np)
	if err != nil {
		return err
	}
	d.SetId(project.ID)
	return resourceQuattroProjectRead(d, meta)
}

func resourceQuattroProjectRead(d *schema.ResourceData, meta interface{}) (err error) {
	defer func() {
		var nErr = rest.GenericOpenAPIError{}
		if errors.As(err, &nErr) {
			err = fmt.Errorf("failed to read project %s: %w", strings.Trim(string(nErr.Body()), "\n "), err)

		}
	}()

	p, err := getConfigFromMeta(meta)
	if err != nil {
		return err
	}

	ctx := p.GetContext()
	project, _, err := p.Client.ProjectsApi.GetByID(ctx, d.Id())
	if err != nil {
		return err
	}
	d.Set(pName, project.Name)

	prof := project.Profile
	pData := map[string]interface{}{
		pAddress:            prof.Address,
		pCompany:            prof.Company,
		pEmail:              prof.Email,
		pEmailVerified:      prof.EmailVerified,
		pPhoneNumber:        prof.PhoneNumber,
		pPhoneVerified:      prof.PhoneVerified,
		pProjectDescription: prof.TeamDesc,
		pProjectName:        prof.TeamName,
	}

	if err = d.Set(pProfile, pData); err != nil {
		return err
	}

	lim := project.Limits
	lData := map[string]interface{}{
		pHosts:           int(lim.Hosts),
		pVolumes:         int(lim.Volumes),
		pVolumeCapacity:  float64(lim.VolumeCapacity),
		pPrivateNetworks: int(lim.PrivateNetworks),
	}

	if err = d.Set(pLimits, lData); err != nil {
		return err
	}
	return nil
}

func resourceQuattroProjectUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	defer func() {
		var nErr = rest.GenericOpenAPIError{}
		if errors.As(err, &nErr) {
			err = fmt.Errorf("failed to update project %s: %w", strings.Trim(string(nErr.Body()), "\n "), err)

		}
	}()

	//p := meta.(*Config)
	return resourceQuattroProjectRead(d, meta)
}

func resourceQuattroProjectDelete(d *schema.ResourceData, meta interface{}) (err error) {
	defer func() {
		var nErr = rest.GenericOpenAPIError{}
		if errors.As(err, &nErr) {
			err = fmt.Errorf("failed to delete project %s: %w", strings.Trim(string(nErr.Body()), "\n "), err)

		}
	}()

	p, err := getConfigFromMeta(meta)
	if err != nil {
		return err
	}

	ctx := p.GetContext()
	_, err = p.Client.ProjectsApi.Delete(ctx, d.Id())
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}
