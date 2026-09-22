// Copyright 2026 Truvity B.V.. All rights reserved.
// SPDX-License-Identifier: MIT

// Package chart_test renders the cnpg-cluster chart and asserts the
// profile postures. The defects that matter here
// render as valid YAML — CNPG defaults enablePDB to TRUE, so merely
// omitting the field on devel silently ships a primary PDB with 0
// allowed disruptions that blocks Karpenter node drains — so the tests
// assert the field is emitted explicitly, not that the manifest parses.
package chart_test

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

// cluster is the sliver of the rendered Cluster these tests assert on.
type (
	cluster struct {
		Kind string `json:"kind"`
		Spec struct {
			Instances int   `json:"instances"`
			EnablePDB *bool `json:"enablePDB"`
			Affinity  struct {
				NodeSelector        map[string]string `json:"nodeSelector"`
				PodAntiAffinityType string            `json:"podAntiAffinityType"`
			} `json:"affinity"`
		} `json:"spec"`
	}
)

func renderCluster(t *testing.T, sets ...string) cluster {
	t.Helper()

	args := []string{"template", "pg", ".", "-n", "test", "--set", "clusterName=test-pg", "--set", "namespace=test"}
	for _, s := range sets {
		args = append(args, "--set", s)
	}

	out, err := exec.CommandContext(t.Context(), "helm", args...).CombinedOutput()
	require.NoErrorf(t, err, "helm template failed:\n%s", out)

	for _, doc := range strings.Split(string(out), "\n---") {
		var c cluster
		if err := yaml.Unmarshal([]byte(doc), &c); err != nil {
			continue // helm's leading comments and empty docs
		}

		if c.Kind == "Cluster" {
			return c
		}
	}

	require.FailNow(t, "render produced no Cluster object:\n"+string(out))

	return cluster{}
}

// TestProfilePDB_ExplicitOnBothPostures: enablePDB must be EMITTED on
// both profiles, never omitted — CNPG's server-side default is true, so
// an omitted field on devel resurrects the drain-blocking primary PDB
// this assertion exists to prevent.
func TestProfilePDB_ExplicitOnBothPostures(t *testing.T) {
	devel := renderCluster(t, "profile=devel")
	require.NotNil(t, devel.Spec.EnablePDB, "devel must emit enablePDB explicitly — omission means CNPG's default (true) and a PDB that blocks node drains")
	assert.False(t, *devel.Spec.EnablePDB, "devel posture is no-PDB: a 1-instance cluster has nowhere to fail over, the PDB only stalls node maintenance")
	assert.Equal(t, 1, devel.Spec.Instances)

	prod := renderCluster(t, "profile=prod")
	require.NotNil(t, prod.Spec.EnablePDB, "prod must emit enablePDB explicitly")
	assert.True(t, *prod.Spec.EnablePDB, "prod posture keeps the PDB: 3 instances make eviction survivable via switchover")
	assert.Equal(t, 3, prod.Spec.Instances)
}

// TestPoolSelector_ProfileAwareDefault: unset databasePool keeps the
// pre-1.1.0 behavior per profile (prod pins "database", devel rides the
// default pools); an explicit value pins BOTH profiles; the empty
// string opts out anywhere. The devel-with-pool case is a devel
// install pinned to a non-interruptible pool.
func TestPoolSelector_ProfileAwareDefault(t *testing.T) {
	prodDefault := renderCluster(t, "profile=prod")
	assert.Equal(t, "database", prodDefault.Spec.Affinity.NodeSelector["karpenter.sh/nodepool"], "prod default must stay the database pool")
	assert.Equal(t, "required", prodDefault.Spec.Affinity.PodAntiAffinityType)

	develDefault := renderCluster(t, "profile=devel")
	assert.Empty(t, develDefault.Spec.Affinity.NodeSelector, "devel default rides the default pools")
	assert.Empty(t, develDefault.Spec.Affinity.PodAntiAffinityType, "anti-affinity is a prod posture")

	develPinned := renderCluster(t, "profile=devel", "scheduling.databasePool=durable")
	assert.Equal(t, "durable", develPinned.Spec.Affinity.NodeSelector["karpenter.sh/nodepool"], "an explicit pool pins the devel profile too")
	assert.Empty(t, develPinned.Spec.Affinity.PodAntiAffinityType, "pinning a pool must not smuggle in the prod anti-affinity")

	prodOptOut := renderCluster(t, "profile=prod", "scheduling.databasePool=")
	assert.Empty(t, prodOptOut.Spec.Affinity.NodeSelector, "empty string opts out on prod")
	assert.Equal(t, "required", prodOptOut.Spec.Affinity.PodAntiAffinityType)
}
