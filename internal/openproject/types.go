// Package openproject provides a client adapter for the OpenProject API.
// Types are re-exported from the generated client at github.com/pinealctx/openproject.
package openproject

import (
	external "github.com/pinealctx/openproject"
)

// Re-export commonly used types from the generated client for convenience.
// Tools and CLI code should import these instead of importing the external package directly.
type (
	// Models
	RootModel              = external.RootModel
	ProjectModel           = external.ProjectModel
	ProjectCollectionModel = external.ProjectCollectionModel
	WorkPackageModel       = external.WorkPackageModel
	WorkPackagesModel      = external.WorkPackagesModel
	WorkPackagePatchModel  = external.WorkPackagePatchModel
	WorkPackageWriteModel  = external.WorkPackageWriteModel
	WorkPackageFormModel   = external.WorkPackageFormModel
	UserModel              = external.UserModel
	UserCollectionModel    = external.UserCollectionModel
	UserCreateModel        = external.UserCreateModel
	MembershipReadModel    = external.MembershipReadModel
	MembershipWriteModel   = external.MembershipWriteModel
	MembershipCollectionModel = external.MembershipCollectionModel
	RoleModel              = external.RoleModel
	RelationReadModel      = external.RelationReadModel
	RelationWriteModel     = external.RelationWriteModel
	RelationCollectionModel = external.RelationCollectionModel
	TimeEntryModel         = external.TimeEntryModel
	TimeEntryCollectionModel = external.TimeEntryCollectionModel
	TimeEntryActivityModel = external.TimeEntryActivityModel
	VersionReadModel       = external.VersionReadModel
	VersionWriteModel      = external.VersionWriteModel
	VersionCollectionModel = external.VersionCollectionModel
	StatusModel            = external.StatusModel
	StatusCollectionModel  = external.StatusCollectionModel
	TypeModel              = external.TypeModel
	TypesByWorkspaceModel  = external.TypesByWorkspaceModel
	PriorityModel          = external.PriorityModel
	PriorityCollectionModel = external.PriorityCollectionModel
	GridReadModel          = external.GridReadModel
	GridWriteModel         = external.GridWriteModel
	GridCollectionModel    = external.GridCollectionModel
	GridWidgetModel        = external.GridWidgetModel
	GroupModel             = external.GroupModel
	GroupCollectionModel   = external.GroupCollectionModel
	NotificationModel      = external.NotificationModel
	NotificationCollectionModel = external.NotificationCollectionModel
	DocumentModel          = external.DocumentModel
	ActivityModel          = external.ActivityModel
	ActivityCommentWriteModel = external.ActivityCommentWriteModel
	Link                   = external.Link
	Formattable            = external.Formattable
	ErrorResponse          = external.ErrorResponse
	AvailableAssigneesModel = external.AvailableAssigneesModel
	WatchersModel          = external.WatchersModel
	PlaceholderUserModel   = external.PlaceholderUserModel
	ConfigurationModel     = external.ConfigurationModel
	PrincipalCollectionModel = external.PrincipalCollectionModel

	// Parameter types
	ListProjectsParams           = external.ListProjectsParams
	ListWorkPackagesParams       = external.ListWorkPackagesParams
	ListUsersParams              = external.ListUsersParams
	ListMembershipsParams        = external.ListMembershipsParams
	ListTimeEntriesParams        = external.ListTimeEntriesParams
	ListRelationsParams          = external.ListRelationsParams
	ListVersionsParams           = external.ListVersionsParams
	ListNotificationsParams      = external.ListNotificationsParams
	ListGridsParams              = external.ListGridsParams
	ListGroupsParams             = external.ListGroupsParams
	ListDocumentsParams          = external.ListDocumentsParams
	GetProjectWorkPackageCollectionParams = external.GetProjectWorkPackageCollectionParams
	UpdateWorkPackageParams      = external.UpdateWorkPackageParams
	CreateProjectWorkPackageParams = external.CreateProjectWorkPackageParams
	ViewWorkPackageParams        = external.ViewWorkPackageParams
	CreateWorkPackageParams      = external.CreateWorkPackageParams
	CommentWorkPackageParams     = external.CommentWorkPackageParams
	ListQueriesParams            = external.ListQueriesParams
	ViewQueryParams              = external.ViewQueryParams
	ListAvailableRelationCandidatesParams = external.ListAvailableRelationCandidatesParams
)
