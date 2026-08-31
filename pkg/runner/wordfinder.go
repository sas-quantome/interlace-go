/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

// constant controlled vocabulary to control automation.
const (
	EnvBasePath  = "ENV_AEE_BASE_PATH"
	EnvDataPath  = "ENV_ALT_DATA_PATH"
	EnvAppsPath  = "ENV_ALT_APPS_PATH"
	UserHomeDir  = "HOME"
	UserLogin    = "USER_LOGIN_NAME"
	UserName     = "USER_NAME"
	LogFile      = "LOG_OUTPUT_FILE"
	ConfigFile   = "CONFIG_FILE"
	QueryTerms   = "QUERY_TERMS"
	OutPath      = "OUT_PATH"
	InpPath      = "INP_PATH"
	InpFile      = "INP_FILE"
	OutFile      = "OUT_FILE"
	IncFile      = "INC_FILE"
	ExcFile      = "EXC_FILE"
	UriInpFile   = "URI_INP_FILE"
	AppProject   = "APP_PROJECT"
	AppName      = "APP_NAME"
	AppAim       = "APP_AIM"
	AppRevision  = "APP_REVISION"
	AppVersion   = "APP_VERSION"
	BaseConcept  = "BASE_CONCEPT"
	AssetNames   = "ASSET_NAMES"
	WorkflowSeq  = "WORKFLOW_SEQUENCE"
	WorkflowID   = "WORKFLOW_ID"
	EventTimeTag = "EVENT_TIME_TAG"
	PrefixName   = "PREFIX_NAME"
	SuffixName   = "SUFFIX_NAME"
	JobStandBy   = "JOB_STANDBY"
	JobRunning   = "JOB_RUNNING"
	JobDone      = "JOB_DONE"
	MaxWorkers   = "MAX_WORKERS"
	MetaInpFile  = "PIPELINE_META"
	SopInpFile   = "PIPELINE_SOP"
	WorkInpFile  = "PIPELINE_WORK"
	DsInpFile    = "DESIGN_PATTERN"
)

// constant base directory paths, everything is systematically organized inside this structure.
const (
	BasePath       = "BASE_PATH"
	DataPath       = "DATA_PATH"
	AppsPath       = "APPS_PATH"
	AgentConfPath  = "AGENT_CONF_PATH"
	AgentExecPath  = "AGENT_EXEC_PATH"
	EventActPath   = "EVENT_ACT_PATH"
	EventArtPath   = "EVENT_ART_PATH"
	EventFunPath   = "EVENT_FUN_PATH"
	EntityCatPath  = "ENTITY_CAT_PATH"
	EntityEnumPath = "ENTITY_ENUM_PATH"
	EntityPropPath = "ENTITY_PROP_PATH"
)

// constant base directory folder names that applications expect immutable.
const (
	ActFN    = "action"
	AgentFN  = "agent"
	ArtFN    = "artifact"
	CatFN    = "catalog"
	ConfigFN = "configuration"
	EntityFN = "entity"
	EnumFN   = "enumeration"
	EventFN  = "event"
	ExecFN   = "execution"
	FuncFN   = "function"
	PropFN   = "property"
)

// ApolloTimeTag represents the timestamp of the Apollo 11 moon landing in the format YYYYMMDDHHMMSS used for testing.
const (
	ApolloTimeTag = "20260101000000" // For test cases only
)
