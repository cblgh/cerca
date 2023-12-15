package constants

const (
	MODLOG_RESETPW = iota
	MODLOG_ADMIN_VETO
	MODLOG_ADMIN_MAKE
	MODLOG_REMOVE_USER
	MODLOG_ADMIN_ADD_USER
	MODLOG_ADMIN_DEMOTE
	/* NOTE: when adding new values, only add them after already existing values! otherwise the existing variables will
	* receive new values */
	// MODLOG_DELETE_VETO
	// MODLOG_DELETE_PROPOSE
	// MODLOG_DELETE_CONFIRM
)
