/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
)

/*

The dispenser's purpose is to provide a consistent file system that does not change and guarantees the reproduction of
data processing events and entity datasets across distributed systems. The dispenser directory structure is designed to
organize data related to agents, entities, and events hierarchically. It follows a specific naming convention to ensure
data organization is unique and clear.

- The agent directory handles agent configuration and pipeline execution scripts.
- The event directory captures workflow logs and Bash scripts generated at runtime.
- The entity directory stores text catalogs, controlled vocabulary enums, and data property binary files.

This system journals all processing events. The outcome compounds and establishes a knowledge hub for downstream analysis.

<interlaceBase>/agent
<interlaceBase>/agent/configuration
<interlaceBase>/agent/execution/<yyyymm>
<interlaceBase>/agent/execution/<yyyymm>/<ddhhmmss>
<interlaceBase>/event
<interlaceBase>/event/<conceptTerm>
<interlaceBase>/event/<conceptTerm>/<workflowID>
<interlaceBase>/event/<conceptTerm>/<workflowID>/<yyyymmddhhmmss>
<interlaceBase>/event/<conceptTerm>/<workflowID>/<yyyymmddhhmmss>/action
<interlaceBase>/event/<conceptTerm>/<workflowID>/<yyyymmddhhmmss>/function
<interlaceBase>/event/<conceptTerm>/<workflowID>/<yyyymmddhhmmss>/artifact
<interlaceBase>/entity
<interlaceBase>/entity/<conceptTerm>
<interlaceBase>/entity/<conceptTerm>/<assetTerm>
<interlaceBase>/entity/<conceptTerm>/<assetTerm>/<yyyymmddhhmmss>
<interlaceBase>/entity/<conceptTerm>/<assetTerm>/<yyyymmddhhmmss>/catalog
<interlaceBase>/entity/<conceptTerm>/<assetTerm>/<yyyymmddhhmmss>/enumeration
<interlaceBase>/entity/<conceptTerm>/<assetTerm>/<yyyymmddhhmmss>/property

*/

// BrokerNamespace represents a unique identifier for an agent manager.
type BrokerNamespace string

// DispenserNamespace represents a unique identifier for a file dispenser.
type DispenserNamespace string

// Dispenser represents a thread-safe structure for managing agents, entities, and events.
type Dispenser struct {
	mu       sync.RWMutex
	BasePath string
	Agent    map[BrokerNamespace]*AgentFiller
	Entity   map[DispenserNamespace]*EntityFiller
	Event    map[DispenserNamespace]*EventFiller
}

// AgentFiller directs agent-specific configuration and execution details.
type AgentFiller struct {
	Name          string
	TimeTag       string
	Execution     []string
	Configuration []string
}

// EntityFiller manages entity-specific dataset catalogs and properties.
type EntityFiller struct {
	Concept     string
	Asset       string
	TimeTag     string
	Catalog     []string
	Enumeration []string
	Property    []DataPartition
}

// EventFiller captures event-specific data processing actions and artifacts.
type EventFiller struct {
	Concept  string
	Workflow string
	TimeTag  string
	Action   []string
	Function []string
	Artifact []string
}

// DataPartition stores data properties into manageable segments.
type DataPartition struct {
	cid string // Content ID (Segment)
	pid string // Partition ID (Fragment)
	bfn string // Binary Filename
}

// NewDispenser creates and initializes a new Dispenser instance.
// The basePath can be empty but must be set before serving.
func NewDispenser(basePath string) *Dispenser {
	return &Dispenser{
		BasePath: basePath,
		Agent:    make(map[BrokerNamespace]*AgentFiller),
		Entity:   make(map[DispenserNamespace]*EntityFiller),
		Event:    make(map[DispenserNamespace]*EventFiller),
	}
}

// ManageNamespace creates a BrokerNamespace from the agentName and timeTag, validating both inputs for correctness.
func ManageNamespace(agentName, timeTag string) (BrokerNamespace, error) {
	if err := ValidateManager(agentName, timeTag); err != nil {
		return "", err
	}
	return BrokerNamespace(agentName + ":" + timeTag), nil
}

// DispenseNamespace creates a DispenserNamespace from the entityConceptName, assetIdentityName, timeTag, and mustBeAppID, validating inputs for correctness.
func DispenseNamespace(entityConceptName, assetIdentityName, timeTag string, mustBeAppID bool) (DispenserNamespace, error) {
	if err := ValidateDispenser(entityConceptName, timeTag, assetIdentityName, mustBeAppID); err != nil {
		return "", err
	}
	return DispenserNamespace(entityConceptName + ":" + assetIdentityName + ":" + timeTag), nil
}

// AddAgent adds an agent to the dispenser, creating a namespace and initializing the agent filler if it doesn't exist.
func (d *Dispenser) AddAgent(agentName, timeTag string) error {
	agentNS, err := ManageNamespace(agentName, timeTag)
	if err != nil {
		return err
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	// Init if not exists (Upsert logic)
	if _, exists := d.Agent[agentNS]; !exists {
		d.Agent[agentNS] = &AgentFiller{
			Name:    agentName,
			TimeTag: timeTag,
		}
	}
	return nil
}

// AddEntity adds an entity to the dispenser, creating a namespace and initializing the entity filler if it doesn't exist.
func (d *Dispenser) AddEntity(concept, asset, timeTag string) error {
	ns, err := DispenseNamespace(concept, asset, timeTag, false)
	if err != nil {
		return err
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.Entity[ns]; !exists {
		d.Entity[ns] = &EntityFiller{
			Concept: concept,
			Asset:   asset,
			TimeTag: timeTag,
		}
	}
	return nil
}

// AddEvent adds an event to the dispenser, creating a namespace and initializing the event filler if it doesn't exist.
func (d *Dispenser) AddEvent(concept, workflow, timeTag string) error {
	ns, err := DispenseNamespace(concept, workflow, timeTag, true)
	if err != nil {
		return err
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.Event[ns]; !exists {
		d.Event[ns] = &EventFiller{
			Concept:  concept,
			Workflow: workflow,
			TimeTag:  timeTag,
		}
	}
	return nil
}

// AddConfigurationFile manually adds a file to the Agent's configuration list.
func (d *Dispenser) AddConfigurationFile(ns BrokerNamespace, fileName string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	// We only add if the agent has been initialized via AddAgent
	if agent, exists := d.Agent[ns]; exists {
		agent.Configuration = append(agent.Configuration, fileName)
	}
}

// AddExecutionFile manually adds a file to the Agent's execution list.
func (d *Dispenser) AddExecutionFile(ns BrokerNamespace, fileName string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if agent, exists := d.Agent[ns]; exists {
		agent.Execution = append(agent.Execution, fileName)
	}
}

// AddCatalogFile manually adds a file to the Entity's catalog list.
func (d *Dispenser) AddCatalogFile(ns DispenserNamespace, fileName string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if entity, exists := d.Entity[ns]; exists {
		entity.Catalog = append(entity.Catalog, fileName)
	}
}

// AddEnumerationFile manually adds a file to the Entity's enumeration list.
func (d *Dispenser) AddEnumerationFile(ns DispenserNamespace, fileName string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if entity, exists := d.Entity[ns]; exists {
		entity.Enumeration = append(entity.Enumeration, fileName)
	}
}

// AddPropertyFile manually adds a file to the Entity's property list.
// It correctly handles the cid (segment) and pid (partition) logic.
func (d *Dispenser) AddPropertyFile(ns DispenserNamespace, cid, pid, fileName string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if entity, exists := d.Entity[ns]; exists {
		p := DataPartition{
			cid: cid,
			pid: pid,
			bfn: fileName,
		}
		entity.Property = append(entity.Property, p)
	}
}

// AddActionFile manually adds a file to the Event's action list.
func (d *Dispenser) AddActionFile(ns DispenserNamespace, fileName string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if event, exists := d.Event[ns]; exists {
		event.Action = append(event.Action, fileName)
	}
}

// AddFunctionFile manually adds a file to the Event's function list.
func (d *Dispenser) AddFunctionFile(ns DispenserNamespace, fileName string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if event, exists := d.Event[ns]; exists {
		event.Function = append(event.Function, fileName)
	}
}

// AddArtifactFile manually adds a file to the Event's artifact (artifact) list.
func (d *Dispenser) AddArtifactFile(ns DispenserNamespace, fileName string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if event, exists := d.Event[ns]; exists {
		event.Artifact = append(event.Artifact, fileName)
	}
}

// --- Path Generators (Read-Locked) ---

func (d *Dispenser) AgentPath() string  { return filepath.Join(d.BasePath, AgentFN) }
func (d *Dispenser) EntityPath() string { return filepath.Join(d.BasePath, EntityFN) }
func (d *Dispenser) EventPath() string  { return filepath.Join(d.BasePath, EventFN) }

// Helper to safely get Agent and generate a path
func (d *Dispenser) getAgentPath(ns BrokerNamespace, subDirs ...string) string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if a, ok := d.Agent[ns]; ok {
		parts := []string{d.AgentPath()}
		// Add custom logic for Execution structure which uses time splitting
		if len(subDirs) > 0 && subDirs[0] == ExecFN {
			parts = append(parts, ExecFN, a.TimeTag[:6], a.TimeTag[6:])
		} else {
			parts = append(parts, subDirs...)
		}
		return filepath.Join(parts...)
	}
	return ""
}

// Helper to safely get Entity and generate a path
func (d *Dispenser) getEntityPath(ns DispenserNamespace, subDirs ...string) string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if e, ok := d.Entity[ns]; ok {
		// Base: <interlaceBase>/entity/<concept>/<asset>/<timetag>
		base := filepath.Join(d.EntityPath(), e.Concept, e.Asset, e.TimeTag)
		return filepath.Join(append([]string{base}, subDirs...)...)
	}
	return ""
}

// Helper to safely get Event and generate a path
func (d *Dispenser) getEventPath(ns DispenserNamespace, subDirs ...string) string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if e, ok := d.Event[ns]; ok {
		// Base: <interlaceBase>/event/<concept>/<workflow>/<timetag>
		base := filepath.Join(d.EventPath(), e.Concept, e.Workflow, e.TimeTag)
		return filepath.Join(append([]string{base}, subDirs...)...)
	}
	return ""
}

// -- Specific Path Accessors --

func (d *Dispenser) AgentConfigurationPath() string { return filepath.Join(d.AgentPath(), ConfigFN) }

func (d *Dispenser) AgentExecutionPath(ns BrokerNamespace) string {
	return d.getAgentPath(ns, ExecFN)
}

func (d *Dispenser) EntityCatalogPath(ns DispenserNamespace) string {
	return d.getEntityPath(ns, CatFN)
}

func (d *Dispenser) EntityEnumerationPath(ns DispenserNamespace) string {
	return d.getEntityPath(ns, EnumFN)
}

func (d *Dispenser) EntityPropertyPath(ns DispenserNamespace) string {
	return d.getEntityPath(ns, PropFN)
}

func (d *Dispenser) EventActionPath(ns DispenserNamespace) string {
	return d.getEventPath(ns, ActFN)
}

func (d *Dispenser) EventFunctionPath(ns DispenserNamespace) string {
	return d.getEventPath(ns, FuncFN)
}

func (d *Dispenser) EventArtifactPath(ns DispenserNamespace) string {
	return d.getEventPath(ns, ArtFN)
}

// AgentConfigurationFiles retrieves the full paths of all configuration files for an agent.
func (d *Dispenser) AgentConfigurationFiles(ns BrokerNamespace) []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// 1. Check if an agent exists
	agent, ok := d.Agent[ns]
	if !ok {
		return nil
	}

	// 2. Get the base directory for these files
	baseDir := d.AgentConfigurationPath()

	// 3. Construct full paths
	files := make([]string, 0, len(agent.Configuration))
	for _, f := range agent.Configuration {
		files = append(files, filepath.Join(baseDir, f))
	}
	return files
}

// AgentExecutionFiles retrieves the full paths of all execution files for an agent.
func (d *Dispenser) AgentExecutionFiles(ns BrokerNamespace) []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	agent, ok := d.Agent[ns]
	if !ok {
		return nil
	}

	// Note: Execution path includes the time tag folder structure (yyyymm/ddhhmmss)
	baseDir := d.AgentExecutionPath(ns)

	files := make([]string, 0, len(agent.Execution))
	for _, f := range agent.Execution {
		files = append(files, filepath.Join(baseDir, f))
	}
	return files
}

// EntityCatalogFiles retrieves the full paths of catalog files.
func (d *Dispenser) EntityCatalogFiles(ns DispenserNamespace) []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	entity, ok := d.Entity[ns]
	if !ok {
		return nil
	}

	baseDir := d.EntityCatalogPath(ns)
	files := make([]string, 0, len(entity.Catalog))
	for _, f := range entity.Catalog {
		files = append(files, filepath.Join(baseDir, f))
	}
	return files
}

// EntityEnumerationFiles retrieves the full paths of enumeration files.
func (d *Dispenser) EntityEnumerationFiles(ns DispenserNamespace) []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	entity, ok := d.Entity[ns]
	if !ok {
		return nil
	}

	baseDir := d.EntityEnumerationPath(ns)
	files := make([]string, 0, len(entity.Enumeration))
	for _, f := range entity.Enumeration {
		files = append(files, filepath.Join(baseDir, f))
	}
	return files
}

// EntityPropertyFiles retrieves the full paths of property files.
// It automatically handles the Content ID (cid) and Partition ID (pid) subdirectories.
func (d *Dispenser) EntityPropertyFiles(ns DispenserNamespace) []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	entity, ok := d.Entity[ns]
	if !ok {
		return nil
	}

	baseDir := d.EntityPropertyPath(ns)
	files := make([]string, 0, len(entity.Property))

	for _, p := range entity.Property {
		// Construct a path based on partition depth:
		// 1. <interlaceBase>/<cid>/<pid>/<file>
		// 2. <interlaceBase>/<cid>/<file>
		// 3. <interlaceBase>/<file>

		var fullPath string
		if p.cid != "" {
			if p.pid != "" {
				fullPath = filepath.Join(baseDir, p.cid, p.pid, p.bfn)
			} else {
				fullPath = filepath.Join(baseDir, p.cid, p.bfn)
			}
		} else {
			fullPath = filepath.Join(baseDir, p.bfn)
		}
		files = append(files, fullPath)
	}
	return files
}

// EventActionFiles retrieves the full paths of action logs.
func (d *Dispenser) EventActionFiles(ns DispenserNamespace) []string {
	return d.getEventFiles(ns, d.EventActionPath, func(e *EventFiller) []string { return e.Action })
}

// EventFunctionFiles retrieves the full paths of function scripts.
func (d *Dispenser) EventFunctionFiles(ns DispenserNamespace) []string {
	return d.getEventFiles(ns, d.EventFunctionPath, func(e *EventFiller) []string { return e.Function })
}

// EventArtifactFiles retrieves the full paths of artifact products.
func (d *Dispenser) EventArtifactFiles(ns DispenserNamespace) []string {
	return d.getEventFiles(ns, d.EventArtifactPath, func(e *EventFiller) []string { return e.Artifact })
}

// Helper to reduce duplication in Event getters
func (d *Dispenser) getEventFiles(ns DispenserNamespace, pathFunc func(DispenserNamespace) string, listSelector func(*EventFiller) []string) []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	event, ok := d.Event[ns]
	if !ok {
		return nil
	}

	baseDir := pathFunc(ns)
	rawList := listSelector(event)

	files := make([]string, 0, len(rawList))
	for _, f := range rawList {
		files = append(files, filepath.Join(baseDir, f))
	}
	return files
}

// --- Directory Creation ---

func (d *Dispenser) createDirs(pathStr string) error {
	return os.MkdirAll(pathStr, 0755)
}

// CreateAgentDirectories ensures that the necessary directory structure exists for the given BrokerNamespace.
func (d *Dispenser) CreateAgentDirectories(ns BrokerNamespace) error {
	if err := d.createDirs(d.AgentConfigurationPath()); err != nil {
		return err
	}
	return d.createDirs(d.AgentExecutionPath(ns))
}

// CreateEventDirectories ensures that the necessary directory structure exists for the given DispenserNamespace.
func (d *Dispenser) CreateEventDirectories(ns DispenserNamespace) error {
	for _, p := range []string{d.EventActionPath(ns), d.EventFunctionPath(ns), d.EventArtifactPath(ns)} {
		if err := d.createDirs(p); err != nil {
			return err
		}
	}
	return nil
}

// CreateEntityDirectories ensures that the necessary directory structure exists for the given DispenserNamespace.
func (d *Dispenser) CreateEntityDirectories(ns DispenserNamespace) error {
	for _, p := range []string{d.EntityCatalogPath(ns), d.EntityEnumerationPath(ns), d.EntityPropertyPath(ns)} {
		if err := d.createDirs(p); err != nil {
			return err
		}
	}
	return nil
}

// --- File Collection Logic ---

// genericCollector scans a directory and executes a callback for each matching file.
func (d *Dispenser) genericCollector(dirPath, pattern string, onMatch func(string)) error {
	if dirPath == "" {
		return nil
	}
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		// Depending on requirements, we might return erring or just nil if dir doesn't exist.
		// Returning err is safer.
		return err
	}
	for _, e := range entries {
		if !e.IsDir() && strings.Contains(e.Name(), pattern) {
			onMatch(e.Name())
		}
	}
	return nil
}

// CollectAgentConfigurationFiles scans and adds config files.
func (d *Dispenser) CollectAgentConfigurationFiles(ns BrokerNamespace, pattern string) error {
	return d.genericCollector(d.AgentConfigurationPath(), pattern, func(f string) {
		d.mu.Lock()
		defer d.mu.Unlock()
		if a := d.Agent[ns]; a != nil {
			a.Configuration = append(a.Configuration, f)
		}
	})
}

// CollectAgentExecutionFiles scans and adds execution files.
func (d *Dispenser) CollectAgentExecutionFiles(ns BrokerNamespace, pattern string) error {
	return d.genericCollector(d.AgentExecutionPath(ns), pattern, func(f string) {
		d.mu.Lock()
		defer d.mu.Unlock()
		if a := d.Agent[ns]; a != nil {
			a.Execution = append(a.Execution, f)
		}
	})
}

// CollectEventFiles is a helper to reduce duplication for Actions, Functions, Artifacts.
func (d *Dispenser) collectEventFiles(ns DispenserNamespace, pattern string, pathFunc func(DispenserNamespace) string, listSelector func(*EventFiller) *[]string) error {
	return d.genericCollector(pathFunc(ns), pattern, func(f string) {
		d.mu.Lock()
		defer d.mu.Unlock()
		if e := d.Event[ns]; e != nil {
			l := listSelector(e)
			*l = append(*l, f)
		}
	})
}

func (d *Dispenser) CollectEventActionFiles(ns DispenserNamespace, pat string) error {
	return d.collectEventFiles(ns, pat, d.EventActionPath, func(e *EventFiller) *[]string { return &e.Action })
}

func (d *Dispenser) CollectEventFunctionFiles(ns DispenserNamespace, pat string) error {
	return d.collectEventFiles(ns, pat, d.EventFunctionPath, func(e *EventFiller) *[]string { return &e.Function })
}

func (d *Dispenser) CollectEventArtifactFiles(ns DispenserNamespace, pat string) error {
	return d.collectEventFiles(ns, pat, d.EventArtifactPath, func(e *EventFiller) *[]string { return &e.Artifact })
}

func (d *Dispenser) CollectEntityCatalogFiles(ns DispenserNamespace, pat string) error {
	return d.genericCollector(d.EntityCatalogPath(ns), pat, func(f string) {
		d.mu.Lock()
		defer d.mu.Unlock()
		if e := d.Entity[ns]; e != nil {
			e.Catalog = append(e.Catalog, f)
		}
	})
}

func (d *Dispenser) CollectEntityEnumerationFiles(ns DispenserNamespace, pat string) error {
	return d.genericCollector(d.EntityEnumerationPath(ns), pat, func(f string) {
		d.mu.Lock()
		defer d.mu.Unlock()
		if e := d.Entity[ns]; e != nil {
			e.Enumeration = append(e.Enumeration, f)
		}
	})
}

// --- Entity Property Walking (Optimized) ---

// CollectEntityPropertyFiles walks the property directory handling:
// 1. Root level files (no partition)
// 2. CID level files (3-letter folders)
// 3. PID level files (3-letter folders inside CID)
func (d *Dispenser) CollectEntityPropertyFiles(ns DispenserNamespace, pattern, filterCID, filterPID string) error {
	rootPath := d.EntityPropertyPath(ns)
	if rootPath == "" {
		return nil
	}

	// We use a local slice to collect matches to minimize lock contention.
	var partitions []DataPartition

	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return err
	}

	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			// Level 1: CID
			if isPartitionCode(name) && matchFilter(name, filterCID) {
				// We found a valid CID folder, traverse into it
				cidPath := filepath.Join(rootPath, name)
				d.walkCIDDir(name, cidPath, pattern, filterPID, &partitions)
			}
		} else if strings.HasSuffix(name, ".bin") && matchFilter(name, pattern) {
			// Level 0: Root file
			partitions = append(partitions, DataPartition{bfn: name})
		}
	}

	// Bulk append protected by lock
	if len(partitions) > 0 {
		d.mu.Lock()
		if e := d.Entity[ns]; e != nil {
			e.Property = append(e.Property, partitions...)
		}
		d.mu.Unlock()
	}

	return nil
}

func (d *Dispenser) walkCIDDir(cid, pathStr, pattern, filterPID string, results *[]DataPartition) {
	entries, _ := os.ReadDir(pathStr) // Ignore error, treat as empty
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			// Level 2: PID
			if isPartitionCode(name) && matchFilter(name, filterPID) {
				pidPath := filepath.Join(pathStr, name)
				d.walkPIDDir(cid, name, pidPath, pattern, results)
			}
		} else if strings.HasSuffix(name, ".bin") && matchFilter(name, pattern) {
			// File inside the CID folder
			*results = append(*results, DataPartition{cid: cid, bfn: name})
		}
	}
}

func (d *Dispenser) walkPIDDir(cid, pid, pathStr, pattern string, results *[]DataPartition) {
	entries, _ := os.ReadDir(pathStr)
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() && strings.HasSuffix(name, ".bin") && matchFilter(name, pattern) {
			// File inside the PID folder
			*results = append(*results, DataPartition{cid: cid, pid: pid, bfn: name})
		}
	}
}

// CoWrAssetParams returns the parameters to open the Cowriter.
func (d *Dispenser) CoWrAssetParams(eventNS DispenserNamespace, assetName string) (string, string, bool, bool) {
	return d.EventArtifactPath(eventNS), assetName + "_*.tsv.gz", true, true
}

// SwapAssetPath returns the path with assetName replaced by queryAssetName.
func (d *Dispenser) SwapAssetPath(entityNS DispenserNamespace, assetName, queryAssetName string) string {
	return strings.Replace(d.EntityPropertyPath(entityNS), assetName, queryAssetName, -1)
}

// matchFilter checks if the given name contains the specified pattern or if the pattern is empty, returning true if so.
func matchFilter(name, pattern string) bool {
	return pattern == "" || strings.Contains(name, pattern)
}

// isPartitionCode replaces "CheckAllAreLowerCased(name) == 3"
func isPartitionCode(s string) bool {
	if len(s) != 3 {
		return false
	}
	for _, r := range s {
		if !unicode.IsLower(r) {
			return false
		}
	}
	return true
}
