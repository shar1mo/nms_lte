package netconf

const (
	NC_RPC_UNKNOWN = iota
	NC_RPC_ACT_GENERIC

	/* ietf-netconf */
	NC_RPC_GETCONFIG
	NC_RPC_EDIT
	NC_RPC_COPY
	NC_RPC_DELETE
	NC_RPC_LOCK
	NC_RPC_UNLOCK
	NC_RPC_GET
	NC_RPC_KILL
	NC_RPC_COMMIT
	NC_RPC_DISCARD
	NC_RPC_CANCEL
	NC_RPC_VALIDATE

	/* ietf-netconf-monitoring */
	NC_RPC_GETSCHEMA

	/* notifications */
	NC_RPC_SUBSCRIBE

	/* ietf-netconf-nmda */
	NC_RPC_GETDATA
	NC_RPC_EDITDATA

	/* ietf-subscribed-notifications */
	NC_RPC_ESTABLISHSUB
	NC_RPC_MODIFYSUB
	NC_RPC_DELETESUB
	NC_RPC_KILLSUB

	/* ietf-yang-push */
	NC_RPC_ESTABLISHPUSH
	NC_RPC_MODIFYPUSH
	NC_RPC_RESYNCSUB
)

const (
	NC_DATASTORE_ERROR     = iota /**< error state of functions returning the datastore type */
	NC_DATASTORE_CONFIG           /**< value describing that the datastore is set as config */
	NC_DATASTORE_URL              /**< value describing that the datastore data should be given from the URL */
	NC_DATASTORE_RUNNING          /**< base NETCONF's datastore containing the current device configuration */
	NC_DATASTORE_STARTUP          /**< separated startup datastore as defined in Distinct Startup Capability */
	NC_DATASTORE_CANDIDATE        /**< separated working datastore as defined in Candidate Configuration Capability */
)
