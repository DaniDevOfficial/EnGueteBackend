package roles

import "log"

const (
	CanEditMeal        = "can_edit_meal"
	CanDeleteMeal      = "can_delete_meal"
	CanCreateMeal      = "can_create_meal"
	CanChangeMealFlags = "can_change_meal_flags"

	CanUpdateGroup = "can_update_group"
	CanDeleteGroup = "can_delete_group"

	CanBanUsers  = "can_ban_users"
	CanUnbanUser = "can_unban_user"
	CanKickUsers = "can_kick_users"

	CanCreateInviteLinks = "can_create_invite_links"
	CanVoidInviteLinks   = "can_void_invite_links"

	CanForceOptIn        = "can_force_opt_in"
	CanSendNotifications = "can_send_notifications"

	CanPromoteToAdmins   = "can_promote_to_admins"
	CanDemoteFromAdmins  = "can_demote_from_admins"
	CanPromoteToManager  = "can_promote_to_manager"
	CanDemoteFromManager = "can_demote_from_manager"
)

const (
	AdminRole   = "admin"
	ManagerRole = "manager"
	MemberRole  = "member"
)

var RolePermissions = map[string]map[string]bool{
	CanEditMeal:        {AdminRole: true, ManagerRole: true, MemberRole: false},
	CanDeleteMeal:      {AdminRole: true, ManagerRole: true, MemberRole: false},
	CanCreateMeal:      {AdminRole: true, ManagerRole: true, MemberRole: false},
	CanChangeMealFlags: {AdminRole: true, ManagerRole: true, MemberRole: false},

	CanUpdateGroup: {AdminRole: true, ManagerRole: true, MemberRole: false},
	CanDeleteGroup: {AdminRole: true, ManagerRole: false, MemberRole: false},

	CanBanUsers:  {AdminRole: true, ManagerRole: false, MemberRole: false},
	CanKickUsers: {AdminRole: true, ManagerRole: true, MemberRole: false},
	CanUnbanUser: {AdminRole: true, ManagerRole: false, MemberRole: false},

	CanCreateInviteLinks: {AdminRole: true, ManagerRole: false, MemberRole: false},
	CanVoidInviteLinks:   {AdminRole: true, ManagerRole: false, MemberRole: false},

	CanForceOptIn:        {AdminRole: true, ManagerRole: false, MemberRole: false},
	CanSendNotifications: {AdminRole: true, ManagerRole: true, MemberRole: false},

	CanPromoteToAdmins:   {AdminRole: true, ManagerRole: false, MemberRole: false},
	CanDemoteFromAdmins:  {AdminRole: true, ManagerRole: false, MemberRole: false},
	CanPromoteToManager:  {AdminRole: true, ManagerRole: false, MemberRole: false},
	CanDemoteFromManager: {AdminRole: true, ManagerRole: false, MemberRole: false},
}

func CanPerformAction(roles []string, action string) bool {
	canDoSpecificAction := RolePermissions[action]
	for _, role := range roles {
		if canDoSpecificAction[role] == true {
			return true
		}
	}
	log.Println("Cannot perform action:", action, "with roles:", roles)
	return false
}

func GetAllRoleRightsForARole(role string) []string {
	var allowedActions []string
	for _, action := range RolePermissions {
		if action[role] == true {
			allowedActions = append(allowedActions, role)
		}
	}
	return allowedActions
}
