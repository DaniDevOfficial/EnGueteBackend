package roles

const (
	CanUpdateMeal      = "can_update_meal"
	CanDeleteMeal      = "can_delete_meal"
	CanCreateMeal      = "can_create_meal"
	CanChangeMealFlags = "can_change_meal_flags"

	CanForceMealPreferenceAndCooking = "can_force_meal_preference_and_cooking"

	CanUpdateGroup = "can_update_group"
	CanDeleteGroup = "can_delete_group"

	CanBanUsers  = "can_ban_users"
	CanUnbanUser = "can_unban_user"
	CanKickUsers = "can_kick_users"

	CanCreateInviteLinks = "can_create_invite_links"
	CanVoidInviteLinks   = "can_void_invite_links"

	CanSendNotifications = "can_send_notifications"

	CanPromoteToAdmins   = "can_promote_to_admin"
	CanDemoteFromAdmins  = "can_demote_from_admin"
	CanPromoteToManager  = "can_promote_to_manager"
	CanDemoteFromManager = "can_demote_from_manager"
)

const (
	AdminRole   = "admin"
	ManagerRole = "manager"
	MemberRole  = "member"
)

var RolePermissions = map[string]map[string]bool{
	CanUpdateMeal:      {AdminRole: true, ManagerRole: true, MemberRole: false},
	CanDeleteMeal:      {AdminRole: true, ManagerRole: true, MemberRole: false},
	CanCreateMeal:      {AdminRole: true, ManagerRole: true, MemberRole: false},
	CanChangeMealFlags: {AdminRole: true, ManagerRole: true, MemberRole: false},

	CanForceMealPreferenceAndCooking: {AdminRole: true, ManagerRole: true, MemberRole: false},

	CanUpdateGroup: {AdminRole: true, ManagerRole: true, MemberRole: false},
	CanDeleteGroup: {AdminRole: true, ManagerRole: false, MemberRole: false},

	CanBanUsers:  {AdminRole: true, ManagerRole: false, MemberRole: false},
	CanKickUsers: {AdminRole: true, ManagerRole: false, MemberRole: false},
	CanUnbanUser: {AdminRole: true, ManagerRole: false, MemberRole: false},

	CanCreateInviteLinks: {AdminRole: true, ManagerRole: false, MemberRole: false},
	CanVoidInviteLinks:   {AdminRole: true, ManagerRole: false, MemberRole: false},

	CanSendNotifications: {AdminRole: true, ManagerRole: true, MemberRole: false},

	CanPromoteToAdmins:   {AdminRole: true, ManagerRole: false, MemberRole: false},
	CanDemoteFromAdmins:  {AdminRole: true, ManagerRole: false, MemberRole: false},
	CanPromoteToManager:  {AdminRole: true, ManagerRole: false, MemberRole: false},
	CanDemoteFromManager: {AdminRole: true, ManagerRole: false, MemberRole: false},
}

func CanPerformAction(roles []string, action string) bool {
	canDoSpecificAction := RolePermissions[action]
	if canDoSpecificAction == nil {
		return false
	}
	for _, role := range roles {
		if canDoSpecificAction[role] == true {
			return true
		}
	}
	return false
}

func GetConstViaString(roleName string) string {
	switch roleName {
	case "admin":
		return AdminRole
	case "manager":
		return ManagerRole
	case "member":
		return MemberRole
	default:
		return ""
	}
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

func GetAllAllowedActionsForRoles(roles []string) []string {
	var allowedActions []string
	for permission := range RolePermissions {
		if CanPerformAction(roles, permission) {
			allowedActions = append(allowedActions, permission)
		}
	}
	return allowedActions
}
