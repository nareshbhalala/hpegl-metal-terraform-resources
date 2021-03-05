// Copyright (c) 2016-2020 Hewlett Packard Enterprise Development LP.

package quake

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	rest "github.com/quattronetworks/quake-client/v1/pkg/client"
)

const (
	sshKeyName   = "name"
	sshPublicKey = "public_key"
)

func sshKeySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		sshKeyName: {
			Type:     schema.TypeString,
			Required: true,
		},

		sshPublicKey: {
			Type:     schema.TypeString,
			Required: true,
		},
	}
}

func SshKeyResource() *schema.Resource {
	return &schema.Resource{
		Create: resourceQuakeSSHKeyCreate,
		Read:   resourceQuakeSSHKeyRead,
		Update: resourceQuakeSSHKeyUpdate,
		Delete: resourceQuakeSSHKeyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: sshKeySchema(),
	}
}

func resourceQuakeSSHKeyCreate(d *schema.ResourceData, meta interface{}) (err error) {
	defer func() {
		var nErr = rest.GenericOpenAPIError{}
		if errors.As(err, &nErr) {
			err = fmt.Errorf("failed to create ssh_key %s: %w", strings.Trim(string(nErr.Body()), "\n "), err)

		}
	}()

	p, err := getConfigFromMeta(meta)
	if err != nil {
		return err
	}
	r := rest.NewSshKey{
		Name: d.Get(sshKeyName).(string),
		Key:  d.Get(sshPublicKey).(string),
	}
	key, _, err := p.Client.SshkeysApi.Add(p.Context, r)
	if err != nil {
		return err
	}
	d.SetId(key.ID)
	if err = p.RefreshAvailableResources(); err != nil {
		return err
	}
	return resourceQuakeSSHKeyRead(d, meta)
}

func resourceQuakeSSHKeyRead(d *schema.ResourceData, meta interface{}) (err error) {
	defer func() {
		var nErr = rest.GenericOpenAPIError{}
		if errors.As(err, &nErr) {
			err = fmt.Errorf("failed to read ssh_key %s: %w", strings.Trim(string(nErr.Body()), "\n "), err)

		}
	}()

	p, err := getConfigFromMeta(meta)
	if err != nil {
		return err
	}
	ssh, _, err := p.Client.SshkeysApi.GetByID(p.Context, d.Id())
	if err != nil {
		return err
	}
	d.Set(sshKeyName, ssh.Name)
	d.Set(sshPublicKey, ssh.Key)
	return nil
}

func resourceQuakeSSHKeyUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	defer func() {
		var nErr = rest.GenericOpenAPIError{}
		if errors.As(err, &nErr) {
			err = fmt.Errorf("failed to update ssh_key %s: %w", strings.Trim(string(nErr.Body()), "\n "), err)

		}
	}()

	p, err := getConfigFromMeta(meta)
	if err != nil {
		return err
	}
	// Read existing
	ssh, _, err := p.Client.SshkeysApi.GetByID(p.Context, d.Id())
	if err != nil {
		return err
	}
	// Modify
	if name, ok := d.Get(sshKeyName).(string); ok && name != "" {
		ssh.Name = name
	}
	if public, ok := d.Get(sshPublicKey).(string); ok && public != "" {
		ssh.Key = public
	}
	// Update
	_, _, err = p.Client.SshkeysApi.Update(p.Context, ssh.ID, ssh)
	if err != nil {
		return err
	}
	return resourceQuakeSSHKeyRead(d, meta)
}

func resourceQuakeSSHKeyDelete(d *schema.ResourceData, meta interface{}) (err error) {
	defer func() {
		var nErr = rest.GenericOpenAPIError{}
		if errors.As(err, &nErr) {
			err = fmt.Errorf("failed to delete ssh_key %s: %w", strings.Trim(string(nErr.Body()), "\n "), err)

		}
	}()

	p, err := getConfigFromMeta(meta)
	if err != nil {
		return err
	}
	_, err = p.Client.SshkeysApi.Delete(p.Context, d.Id())
	if err != nil {
		return err
	}
	d.SetId("")
	return p.RefreshAvailableResources()
}
