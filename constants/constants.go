package constants
import "time"

const (
	MODLOG_RESETPW = iota
	MODLOG_ADMIN_VETO // vetoing a proposal
	MODLOG_ADMIN_MAKE // make an admin
	MODLOG_REMOVE_USER // remove a user
	MODLOG_ADMIN_ADD_USER // add a new user
	MODLOG_ADMIN_DEMOTE // demote an admin back to a normal user
	MODLOG_ADMIN_CONFIRM // confirming a proposal
	MODLOG_ADMIN_PROPOSE_DEMOTE_ADMIN
	MODLOG_ADMIN_PROPOSE_MAKE_ADMIN
	MODLOG_ADMIN_PROPOSE_REMOVE_USER
	/* NOTE: when adding new values, only add them after already existing values! otherwise the existing variables will
	* receive new values which affects the stored values in table moderation_log */
)

const PROPOSAL_VETO = false
const PROPOSAL_CONFIRM = true
const PROPOSAL_SELF_CONFIRMATION_WAIT = time.Hour * 24 * 7 /* 1 week */
