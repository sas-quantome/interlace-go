/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"path"
	"testing"

	"github.com/sas-quantome/interlace-go/pkg/tester"
	"github.com/stretchr/testify/require"
)

func TestDispenser_NewDispenser(t *testing.T) {
	dispenser := NewDispenser(tester.TestBasePath())
	require.NotNil(t, dispenser)
	require.Equal(t, tester.TestBasePath(), dispenser.BasePath)
}

func TestDispenser_AddAgent(t *testing.T) {
	dispenser, agentNS := createTestAgent(t)
	require.NotNil(t, dispenser.Agent[agentNS])
	require.True(t, ValidTimeTag(dispenser.Agent[agentNS].TimeTag))
}

func TestDispenser_AddEvent(t *testing.T) {
	dispenser, eventNS := createTestEvent(t)
	require.NotNil(t, dispenser.Event[eventNS])
	require.Equal(t, dispenser.Event[eventNS].TimeTag, dispenser.Event[eventNS].TimeTag)
}

func TestDispenser_AddEntity(t *testing.T) {
	dispenser, entityNS := createTestEntity(t)
	require.NotNil(t, dispenser.Entity[entityNS])
	require.True(t, ValidTimeTag(dispenser.Entity[entityNS].TimeTag))
}

func TestDispenser_AgentPath(t *testing.T) {
	dispenser := NewDispenser(tester.TestBasePath())
	timeTag := TimeTagNow()
	require.NoError(t, dispenser.AddAgent("a11408", timeTag))
	require.Equal(t, path.Join(tester.TestBasePath(), AgentFN), dispenser.AgentPath())
}

func TestDispenser_AgentConfigurationPath(t *testing.T) {
	dispenser := NewDispenser(tester.TestBasePath())
	timeTag := TimeTagNow()
	require.NoError(t, dispenser.AddAgent("a11408", timeTag))
	require.Equal(t, path.Join(tester.TestBasePath(), AgentFN, ConfigFN), dispenser.AgentConfigurationPath())
}

func TestDispenser_AgentExecutionPath(t *testing.T) {
	dispenser, agentNS := createTestAgent(t)
	timeTag := dispenser.Agent[agentNS].TimeTag
	require.Equal(t, path.Join(tester.TestBasePath(), AgentFN, ExecFN, timeTag[:6], timeTag[6:]),
		dispenser.AgentExecutionPath(agentNS))
}

func TestDispenser_EntityPath(t *testing.T) {
	dispenser := NewDispenser(tester.TestBasePath())
	timeTag := TimeTagNow()
	require.NoError(t, dispenser.AddEntity("btc", "btc_usd", timeTag))
	require.Equal(t, path.Join(tester.TestBasePath(), EntityFN), dispenser.EntityPath())
}

func TestDispenser_EntityCatalogPath(t *testing.T) {
	dispenser, entityNS := createTestEntity(t)
	require.Equal(t, path.Join(tester.TestBasePath(), EntityFN, "btc", "btc_usd",
		dispenser.Entity[entityNS].TimeTag, CatFN), dispenser.EntityCatalogPath(entityNS))
}

func TestDispenser_EntityEnumerationPath(t *testing.T) {
	dispenser, entityNS := createTestEntity(t)
	require.Equal(t, path.Join(tester.TestBasePath(), EntityFN, "btc", "btc_usd",
		dispenser.Entity[entityNS].TimeTag, EnumFN), dispenser.EntityEnumerationPath(entityNS))
}

func TestDispenser_EntityPropertyPath(t *testing.T) {
	dispenser, entityNS := createTestEntity(t)
	require.Equal(t, path.Join(tester.TestBasePath(), EntityFN, "btc", "btc_usd",
		dispenser.Entity[entityNS].TimeTag, PropFN), dispenser.EntityPropertyPath(entityNS))
}

func TestDispenser_EventPath(t *testing.T) {
	dispenser := NewDispenser(tester.TestBasePath())
	timeTag := TimeTagNow()
	require.NoError(t, dispenser.AddEvent("btc", "w10101", timeTag))
	require.Equal(t, path.Join(tester.TestBasePath(), EventFN), dispenser.EventPath())
}

func TestDispenser_EventActionPath(t *testing.T) {
	dispenser, eventNS := createTestEvent(t)
	require.Equal(t, path.Join(tester.TestBasePath(), EventFN, "btc", "w10101",
		dispenser.Event[eventNS].TimeTag, ActFN), dispenser.EventActionPath(eventNS))
}

func TestDispenser_EventFunctionPath(t *testing.T) {
	dispenser, eventNS := createTestEvent(t)
	require.Equal(t, path.Join(tester.TestBasePath(), EventFN, "btc", "w10101",
		dispenser.Event[eventNS].TimeTag, FuncFN), dispenser.EventFunctionPath(eventNS))
}

func TestDispenser_EventArtifactPath(t *testing.T) {
	dispenser, eventNS := createTestEvent(t)
	require.Equal(t, path.Join(tester.TestBasePath(), EventFN, "btc", "w10101",
		dispenser.Event[eventNS].TimeTag, ArtFN), dispenser.EventArtifactPath(eventNS))
}

func TestDispenser_AddConfigurationFile(t *testing.T) {
	dispenser, agentNS := createTestAgent(t)
	dispenser.AddConfigurationFile(agentNS, "config.tsv")
	require.Equal(t, 1, len(dispenser.Agent[agentNS].Configuration))
	s := dispenser.AgentConfigurationFiles(agentNS)
	require.Equal(t, path.Join(tester.TestBasePath(), AgentFN, ConfigFN, "config.tsv"), s[0])
}

func TestDispenser_AddExecutionFile(t *testing.T) {
	dispenser, agentNS := createTestAgent(t)
	dispenser.AddExecutionFile(agentNS, "w10101")
	require.Equal(t, 1, len(dispenser.Agent[agentNS].Execution))
	s := dispenser.AgentExecutionFiles(agentNS)
	timeTag := dispenser.Agent[agentNS].TimeTag
	require.Equal(t, path.Join(tester.TestBasePath(), AgentFN, ExecFN, timeTag[:6], timeTag[6:], "w10101"), s[0])
}

func TestDispenser_AddCatalogFile(t *testing.T) {
	dispenser, entityNS := createTestEntity(t)
	dispenser.AddCatalogFile(entityNS, "suk_go_term_hash_id_m.gob.gz")
	dispenser.AddCatalogFile(entityNS, "uuk_go_term_text_s.gob.gz")
	require.Equal(t, 2, len(dispenser.Entity[entityNS].Catalog))
	s := dispenser.EntityCatalogFiles(entityNS)
	require.Equal(t, path.Join(dispenser.EntityCatalogPath(entityNS), "suk_go_term_hash_id_m.gob.gz"), s[0])
	require.Equal(t, path.Join(dispenser.EntityCatalogPath(entityNS), "uuk_go_term_text_s.gob.gz"), s[1])
}

func TestDispenser_AddEnumerationFile(t *testing.T) {
	dispenser, entityNS := createTestEntity(t)
	dispenser.AddEnumerationFile(entityNS, "uuk_entrez_category_cv_s.gob.gz")
	require.Equal(t, 1, len(dispenser.Entity[entityNS].Enumeration))
	s := dispenser.EntityEnumerationFiles(entityNS)
	require.Equal(t, path.Join(dispenser.EntityEnumerationPath(entityNS), "uuk_entrez_category_cv_s.gob.gz"), s[0])
}

func TestDispenser_AddPropertyFile(t *testing.T) {
	dispenser, entityNS := createTestEntity(t)
	dispenser.AddPropertyFile(entityNS, "", "", "suk_entrez_gene_id_i.gob.gz")
	dispenser.AddPropertyFile(entityNS, "baa", "", "suk_baa_s.gob.gz")
	dispenser.AddPropertyFile(entityNS, "baa", "aaa", "suk_baa_aaa_s.gob.gz")
	require.Equal(t, 3, len(dispenser.Entity[entityNS].Property))
	s := dispenser.EntityPropertyFiles(entityNS)
	require.Equal(t, path.Join(dispenser.EntityPropertyPath(entityNS), "suk_entrez_gene_id_i.gob.gz"), s[0])
	require.Equal(t, path.Join(dispenser.EntityPropertyPath(entityNS), "baa", "suk_baa_s.gob.gz"), s[1])
	require.Equal(t, path.Join(dispenser.EntityPropertyPath(entityNS), "baa", "aaa", "suk_baa_aaa_s.gob.gz"), s[2])
}

func TestDispenser_AddActionFile(t *testing.T) {
	dispenser, eventNS := createTestEvent(t)
	dispenser.AddActionFile(eventNS, "b10101.log")
	require.Equal(t, 1, len(dispenser.Event[eventNS].Action))
	s := dispenser.EventActionFiles(eventNS)
	require.Equal(t, path.Join(dispenser.EventActionPath(eventNS), "b10101.log"), s[0])
}

func TestDispenser_AddFunctionFile(t *testing.T) {
	dispenser, eventNS := createTestEvent(t)
	dispenser.AddFunctionFile(eventNS, "b10101")
	require.Equal(t, 1, len(dispenser.Event[eventNS].Function))
	s := dispenser.EventFunctionFiles(eventNS)
	require.Equal(t, path.Join(dispenser.EventFunctionPath(eventNS), "b10101"), s[0])
}

func TestDispenser_AddArtifactFile(t *testing.T) {
	dispenser, eventNS := createTestEvent(t)
	dispenser.AddArtifactFile(eventNS, "enetrez_gene_db_18475838743.tsv.gz")
	require.Equal(t, 1, len(dispenser.Event[eventNS].Artifact))
	s := dispenser.EventArtifactFiles(eventNS)
	require.Equal(t, path.Join(dispenser.EventArtifactPath(eventNS), "enetrez_gene_db_18475838743.tsv.gz"), s[0])
}

func TestDispenser_CreateAgentDirectories(t *testing.T) {
	dispenser, agentNS := createTestAgent(t)
	require.NoError(t, dispenser.CreateAgentDirectories(agentNS))
	require.DirExists(t, dispenser.AgentConfigurationPath())
	require.DirExists(t, dispenser.AgentExecutionPath(agentNS))
}

func TestDispenser_CreateEventDirectories(t *testing.T) {
	dispenser, eventNS := createTestEvent(t)
	require.NoError(t, dispenser.CreateEventDirectories(eventNS))
	require.DirExists(t, dispenser.EventActionPath(eventNS))
	require.DirExists(t, dispenser.EventFunctionPath(eventNS))
	require.DirExists(t, dispenser.EventArtifactPath(eventNS))
}

func TestDispenser_CreateEntityDirectories(t *testing.T) {
	require.Equal(t, "ASSET_NAMES", AssetNames)
	require.Equal(t, "PREFIX_NAME", PrefixName)
	require.Equal(t, "JOB_RUNNING", JobRunning)
	require.Equal(t, "ENTITY_CAT_PATH", EntityCatPath)
	require.Equal(t, "ENTITY_ENUM_PATH", EntityEnumPath)
	require.Equal(t, "ENTITY_PROP_PATH", EntityPropPath) // to silence
	dispenser, entityNS := createTestEntity(t)
	require.NoError(t, dispenser.CreateEntityDirectories(entityNS))
	require.DirExists(t, dispenser.EntityCatalogPath(entityNS))
	require.DirExists(t, dispenser.EntityEnumerationPath(entityNS))
	require.DirExists(t, dispenser.EntityPropertyPath(entityNS))
}

func createTestAgent(t *testing.T) (*Dispenser, BrokerNamespace) {
	dispenser := NewDispenser(tester.TestBasePath())
	timeTag := ApolloTimeTag
	require.NoError(t, dispenser.AddAgent("a11408", timeTag))
	agentNS, err := ManageNamespace("a11408", timeTag)
	require.NoError(t, err)
	return dispenser, agentNS
}

func createTestEntity(t *testing.T) (*Dispenser, DispenserNamespace) {
	dispenser := NewDispenser(tester.TestBasePath())
	timeTag := ApolloTimeTag
	require.NoError(t, dispenser.AddEntity("btc", "btc_usd", timeTag))
	entityNS, err := DispenseNamespace("btc", "btc_usd", timeTag, false)
	require.NoError(t, err)
	return dispenser, entityNS
}

func createTestEvent(t *testing.T) (*Dispenser, DispenserNamespace) {
	dispenser := NewDispenser(tester.TestBasePath())
	timeTag := ApolloTimeTag
	require.NoError(t, dispenser.AddEvent("btc", "w10101", timeTag))
	eventNS, err := DispenseNamespace("btc", "w10101", timeTag, true)
	require.NoError(t, err)
	return dispenser, eventNS
}
