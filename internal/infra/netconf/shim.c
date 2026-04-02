#include "shim.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define MODULES_DIR "/home/nqs/nms_rc/third_party/libnetconf2/modules"
#define NCGO_SEND_TIMEOUT_MS 1000
#define NCGO_REPLY_TIMEOUT_MS 10000

static const char *ncgo_nonempty(const char *value);
static int ncgo_prepare_output(char **out_xml, char **out_err);
static int ncgo_fail(char **out_err, const char *msg);
static int ncgo_fail_with_detail(char **out_err, const char *context, const char *detail);
static int ncgo_fail_last(char **out_err, const char *context);
static int ncgo_parse_datastore(const char *name, NC_DATASTORE fallback, NC_DATASTORE *out_ds, char **out_err);
static int ncgo_dispatch_rpc(
    ncgo_client_t *client,
    struct nc_rpc *rpc,
    const char *build_context,
    char **out_xml,
    char **out_err
);
static int ncgo_exec_rpc(
    ncgo_client_t *client,
    struct nc_rpc *rpc,
    char **out_xml,
    char **out_err
);

static const char *ncgo_nonempty(const char *value) {
    if (value && value[0]) {
        return value;
    }

    return NULL;
}

static int ncgo_prepare_output(char **out_xml, char **out_err) {
    if (!out_xml) {
        return ncgo_fail(out_err, "out_xml is required");
    }

    *out_xml = NULL;
    if (out_err) {
        *out_err = NULL;
    }

    return 0;
}

static int ncgo_fail(char **out_err, const char *msg) {
    if (out_err) {
        *out_err = strdup(msg ? msg : "unknown error");
    }

    return -1;
}

static int ncgo_fail_with_detail(char **out_err, const char *context, const char *detail) {
    const char *base = context ? context : "operation failed";

    if (!detail || !detail[0]) {
        return ncgo_fail(out_err, base);
    }

    if (!out_err) {
        return -1;
    }

    size_t base_len = strlen(base);
    size_t detail_len = strlen(detail);
    size_t total_len = base_len + 2 + detail_len + 1;
    char *msg = malloc(total_len);
    if (!msg) {
        return ncgo_fail(out_err, base);
    }

    snprintf(msg, total_len, "%s: %s", base, detail);
    *out_err = msg;
    return -1;
}

static int ncgo_fail_last(char **out_err, const char *context) {
    return ncgo_fail_with_detail(out_err, context, ly_last_errmsg());
}

static int ncgo_parse_datastore(const char *name, NC_DATASTORE fallback, NC_DATASTORE *out_ds, char **out_err) {
    const char *resolved = ncgo_nonempty(name);

    if (!out_ds) {
        return ncgo_fail(out_err, "invalid datastore output");
    }

    if (!resolved) {
        if (fallback != NC_DATASTORE_ERROR) {
            *out_ds = fallback;
            return 0;
        }

        return ncgo_fail(out_err, "datastore is required");
    }

    if (!strcmp(resolved, "candidate")) {
        *out_ds = NC_DATASTORE_CANDIDATE;
    } else if (!strcmp(resolved, "running")) {
        *out_ds = NC_DATASTORE_RUNNING;
    } else if (!strcmp(resolved, "startup")) {
        *out_ds = NC_DATASTORE_STARTUP;
    } else if (!strcmp(resolved, "config")) {
        *out_ds = NC_DATASTORE_CONFIG;
    } else if (!strcmp(resolved, "url")) {
        *out_ds = NC_DATASTORE_URL;
    } else {
        return ncgo_fail(out_err, "invalid datastore, use running/startup/candidate/config/url");
    }

    return 0;
}

int ncgo_client_init(void) {
    nc_client_init();
    nc_verbosity(NC_VERB_WARNING);
    return 0;
}

void ncgo_client_destroy(void) {
    nc_client_destroy();
}

static char *ncgo_password_cb(const char *username, const char *hostname, void *priv) {
    (void)username;
    (void)hostname;

    const char *password = (const char *)priv;
    if (!password) {
        return NULL;
    }

    return strdup(password);
}

int ncgo_connect_ssh(
    const char *host,
    uint16_t port,
    const char *username,
    const char *password,
    const char *schema_searchpath,
    ncgo_client_t **out_client,
    char **out_err
) {
    const char *searchpath = ncgo_nonempty(schema_searchpath);
    ncgo_client_t *client = NULL;

    if (out_err) {
        *out_err = NULL;
    }

    if (!host || !host[0] || !username || !username[0] || !password || !out_client) {
        return ncgo_fail(out_err, "invalid arguments");
    }

    *out_client = NULL;
    if (!searchpath) {
        searchpath = MODULES_DIR;
    }

    client = calloc(1, sizeof(*client));
    if (!client) {
        return ncgo_fail(out_err, "calloc failed");
    }

    client->password_copy = strdup(password);
    if (!client->password_copy) {
        free(client);
        return ncgo_fail(out_err, "strdup failed");
    }

    if (nc_client_set_schema_searchpath(searchpath)) {
        ncgo_close(client);
        return ncgo_fail_last(out_err, "nc_client_set_schema_searchpath failed");
    }

    nc_client_ssh_set_auth_pref(NC_SSH_AUTH_PASSWORD, 4);
    nc_client_ssh_set_username(username);
    nc_client_ssh_set_auth_password_clb(ncgo_password_cb, client->password_copy);
    nc_client_ssh_set_knownhosts_mode(NC_SSH_KNOWNHOSTS_ACCEPT_NEW);

    client->sess = nc_connect_ssh(host, port, NULL);
    if (!client->sess) {
        ncgo_close(client);
        return ncgo_fail_last(out_err, "nc_connect_ssh failed");
    }

    *out_client = client;
    return 0;
}

int ncgo_session_capabilities(
    ncgo_client_t *client,
    char **out_capabilities,
    char **out_err
) {
    const char * const *capabilities;
    size_t total_len = 1;
    int i;
    char *joined;
    char *cursor;

    if (out_err) {
        *out_err = NULL;
    }
    if (!out_capabilities) {
        return ncgo_fail(out_err, "out_capabilities is required");
    }

    *out_capabilities = NULL;
    if (!client || !client->sess) {
        return ncgo_fail(out_err, "client is not connected");
    }

    capabilities = nc_session_get_cpblts(client->sess);
    if (!capabilities) {
        return 0;
    }

    for (i = 0; capabilities[i]; ++i) {
        total_len += strlen(capabilities[i]) + 1;
    }

    joined = malloc(total_len);
    if (!joined) {
        return ncgo_fail(out_err, "malloc failed");
    }

    cursor = joined;
    for (i = 0; capabilities[i]; ++i) {
        size_t capability_len = strlen(capabilities[i]);
        memcpy(cursor, capabilities[i], capability_len);
        cursor += capability_len;
        *cursor++ = '\n';
    }

    if (cursor != joined) {
        cursor--;
    }
    *cursor = '\0';

    *out_capabilities = joined;
    return 0;
}

int ncgo_session_is_alive(ncgo_client_t *client) {
    if (!client || !client->sess) {
        return 0;
    }

    return nc_session_get_status(client->sess) == NC_STATUS_RUNNING;
}

int ncgo_rpc(
    ncgo_client_t *client,
    NC_RPC_TYPE rpc_type,
    const char *payload,
    char **out_xml,
    char **out_err
) {
    const char *content = ncgo_nonempty(payload);

    if (ncgo_prepare_output(out_xml, out_err) != 0) {
        return -1;
    }

    switch (rpc_type) {
    case NC_RPC_ACT_GENERIC: {
        if (!content) {
            return ncgo_fail(out_err, "payload is required for a generic RPC");
        }

        struct nc_rpc *rpc = nc_rpc_act_generic_xml(content, NC_PARAMTYPE_CONST);
        return ncgo_dispatch_rpc(client, rpc, "nc_rpc_act_generic_xml failed", out_xml, out_err);
    }

    case NC_RPC_GET:
        return ncgo_rpc_get(client, content, out_xml, out_err);

    case NC_RPC_GETCONFIG:
        return ncgo_rpc_getconfig(client, "running", content, out_xml, out_err);

    default:
        return ncgo_fail(out_err, "unsupported rpc type");
    }
}

int ncgo_rpc_get(
    ncgo_client_t *client,
    const char *filter,
    char **out_xml,
    char **out_err
) {
    if (ncgo_prepare_output(out_xml, out_err) != 0) {
        return -1;
    }

    struct nc_rpc *rpc = nc_rpc_get(ncgo_nonempty(filter), NC_WD_UNKNOWN, NC_PARAMTYPE_CONST);
    return ncgo_dispatch_rpc(client, rpc, "nc_rpc_get failed", out_xml, out_err);
}

int ncgo_rpc_getconfig(
    ncgo_client_t *client,
    const char *datastore,
    const char *filter,
    char **out_xml,
    char **out_err
) {
    NC_DATASTORE ds = NC_DATASTORE_ERROR;
    if (ncgo_prepare_output(out_xml, out_err) != 0) {
        return -1;
    }

    if (ncgo_parse_datastore(datastore, NC_DATASTORE_RUNNING, &ds, out_err) != 0) {
        return -1;
    }

    struct nc_rpc *rpc = nc_rpc_getconfig(ds, ncgo_nonempty(filter), NC_WD_UNKNOWN, NC_PARAMTYPE_CONST);
    return ncgo_dispatch_rpc(client, rpc, "nc_rpc_getconfig failed", out_xml, out_err);
}

int ncgo_rpc_edit(
    ncgo_client_t *client,
    const char *datastore,
    const char *edit_content,
    char **out_xml,
    char **out_err
) {
    NC_DATASTORE ds = NC_DATASTORE_ERROR;

    if (ncgo_prepare_output(out_xml, out_err) != 0) {
        return -1;
    }

    if (!edit_content || !edit_content[0]) {
        return ncgo_fail(out_err, "edit_content is required");
    }

    if (ncgo_parse_datastore(datastore, NC_DATASTORE_RUNNING, &ds, out_err) != 0) {
        return -1;
    }

    struct nc_rpc *rpc = nc_rpc_edit(
        ds,
        NC_RPC_EDIT_DFLTOP_UNKNOWN,
        NC_RPC_EDIT_TESTOPT_UNKNOWN,
        NC_RPC_EDIT_ERROPT_UNKNOWN,
        edit_content,
        NC_PARAMTYPE_CONST
    );
    return ncgo_dispatch_rpc(client, rpc, "nc_rpc_edit failed", out_xml, out_err);
}

int ncgo_rpc_copy(
    ncgo_client_t *client,
    const char *target,
    const char *url_trg,
    const char *source,
    const char *url_or_config_src,
    char **out_xml,
    char **out_err
) {
    NC_DATASTORE target_ds = NC_DATASTORE_ERROR;
    NC_DATASTORE source_ds = NC_DATASTORE_ERROR;

    if (ncgo_prepare_output(out_xml, out_err) != 0) {
        return -1;
    }

    if (ncgo_parse_datastore(target, NC_DATASTORE_RUNNING, &target_ds, out_err) != 0) {
        return -1;
    }

    if (ncgo_parse_datastore(source, NC_DATASTORE_RUNNING, &source_ds, out_err) != 0) {
        return -1;
    }

    struct nc_rpc *rpc = nc_rpc_copy(
        target_ds,
        ncgo_nonempty(url_trg),
        source_ds,
        ncgo_nonempty(url_or_config_src),
        NC_WD_UNKNOWN,
        NC_PARAMTYPE_CONST
    );
    return ncgo_dispatch_rpc(client, rpc, "nc_rpc_copy failed", out_xml, out_err);
}

int ncgo_rpc_delete(
    ncgo_client_t *client,
    const char *target,
    const char *url,
    char **out_xml,
    char **out_err
) {
    NC_DATASTORE ds = NC_DATASTORE_ERROR;
    if (ncgo_prepare_output(out_xml, out_err) != 0) {
        return -1;
    }

    if (ncgo_parse_datastore(target, NC_DATASTORE_RUNNING, &ds, out_err) != 0) {
        return -1;
    }

    struct nc_rpc *rpc = nc_rpc_delete(ds, ncgo_nonempty(url), NC_PARAMTYPE_CONST);
    return ncgo_dispatch_rpc(client, rpc, "nc_rpc_delete failed", out_xml, out_err);
}

int ncgo_rpc_lock(
    ncgo_client_t *client,
    const char *datastore,
    char **out_xml,
    char **out_err
) {
    NC_DATASTORE ds = NC_DATASTORE_ERROR;
    if (ncgo_prepare_output(out_xml, out_err) != 0) {
        return -1;
    }

    if (ncgo_parse_datastore(datastore, NC_DATASTORE_RUNNING, &ds, out_err) != 0) {
        return -1;
    }

    struct nc_rpc *rpc = nc_rpc_lock(ds);
    return ncgo_dispatch_rpc(client, rpc, "nc_rpc_lock failed", out_xml, out_err);
}

int ncgo_rpc_unlock(
    ncgo_client_t *client,
    const char *datastore,
    char **out_xml,
    char **out_err
) {
    NC_DATASTORE ds = NC_DATASTORE_ERROR;
    if (ncgo_prepare_output(out_xml, out_err) != 0) {
        return -1;
    }

    if (ncgo_parse_datastore(datastore, NC_DATASTORE_RUNNING, &ds, out_err) != 0) {
        return -1;
    }

    struct nc_rpc *rpc = nc_rpc_unlock(ds);
    return ncgo_dispatch_rpc(client, rpc, "nc_rpc_unlock failed", out_xml, out_err);
}

int ncgo_rpc_commit(
    ncgo_client_t *client,
    char **out_xml,
    char **out_err
) {
    if (ncgo_prepare_output(out_xml, out_err) != 0) {
        return -1;
    }

    struct nc_rpc *rpc = nc_rpc_commit(0, 0, NULL, NULL, NC_PARAMTYPE_CONST);
    return ncgo_dispatch_rpc(client, rpc, "nc_rpc_commit failed", out_xml, out_err);
}

int ncgo_rpc_discard(
    ncgo_client_t *client,
    char **out_xml,
    char **out_err
) {
    if (ncgo_prepare_output(out_xml, out_err) != 0) {
        return -1;
    }

    struct nc_rpc *rpc = nc_rpc_discard();
    return ncgo_dispatch_rpc(client, rpc, "nc_rpc_discard failed", out_xml, out_err);
}

int ncgo_rpc_cancel(
    ncgo_client_t *client,
    const char *persist_id,
    char **out_xml,
    char **out_err
) {
    if (ncgo_prepare_output(out_xml, out_err) != 0) {
        return -1;
    }

    struct nc_rpc *rpc = nc_rpc_cancel(ncgo_nonempty(persist_id), NC_PARAMTYPE_CONST);
    return ncgo_dispatch_rpc(client, rpc, "nc_rpc_cancel failed", out_xml, out_err);
}

int ncgo_rpc_validate(
    ncgo_client_t *client,
    const char *source,
    const char *url_or_config,
    char **out_xml,
    char **out_err
) {
    NC_DATASTORE ds = NC_DATASTORE_ERROR;
    if (ncgo_prepare_output(out_xml, out_err) != 0) {
        return -1;
    }

    if (ncgo_parse_datastore(source, NC_DATASTORE_RUNNING, &ds, out_err) != 0) {
        return -1;
    }

    struct nc_rpc *rpc = nc_rpc_validate(ds, ncgo_nonempty(url_or_config), NC_PARAMTYPE_CONST);
    return ncgo_dispatch_rpc(client, rpc, "nc_rpc_validate failed", out_xml, out_err);
}

int ncgo_rpc_getschema(
    ncgo_client_t *client,
    const char *identifier,
    const char *version,
    const char *format,
    char **out_xml,
    char **out_err
) {
    if (ncgo_prepare_output(out_xml, out_err) != 0) {
        return -1;
    }

    if (!identifier || !identifier[0]) {
        return ncgo_fail(out_err, "identifier is required");
    }

    struct nc_rpc *rpc = nc_rpc_getschema(
        identifier,
        ncgo_nonempty(version),
        ncgo_nonempty(format),
        NC_PARAMTYPE_CONST
    );
    return ncgo_dispatch_rpc(client, rpc, "nc_rpc_getschema failed", out_xml, out_err);
}

int ncgo_rpc_subscribe(
    ncgo_client_t *client,
    const char *stream_name,
    const char *filter,
    const char *start_time,
    const char *stop_time,
    char **out_xml,
    char **out_err
) {
    if (ncgo_prepare_output(out_xml, out_err) != 0) {
        return -1;
    }

    struct nc_rpc *rpc = nc_rpc_subscribe(
        ncgo_nonempty(stream_name),
        ncgo_nonempty(filter),
        ncgo_nonempty(start_time),
        ncgo_nonempty(stop_time),
        NC_PARAMTYPE_CONST
    );
    return ncgo_dispatch_rpc(client, rpc, "nc_rpc_subscribe failed", out_xml, out_err);
}

int ncgo_rpc_getdata(
    ncgo_client_t *client,
    const char *datastore,
    const char *filter,
    char **out_xml,
    char **out_err
) {
    if (ncgo_prepare_output(out_xml, out_err) != 0) {
        return -1;
    }

    if (!datastore || !datastore[0]) {
        return ncgo_fail(out_err, "datastore is required");
    }

    struct nc_rpc *rpc = nc_rpc_getdata(
        datastore,
        ncgo_nonempty(filter),
        NULL,
        NULL,
        0,
        0,
        0,
        0,
        NC_WD_UNKNOWN,
        NC_PARAMTYPE_CONST
    );
    return ncgo_dispatch_rpc(client, rpc, "nc_rpc_getdata failed", out_xml, out_err);
}

int ncgo_rpc_editdata(
    ncgo_client_t *client,
    const char *datastore,
    const char *edit_content,
    char **out_xml,
    char **out_err
) {
    if (ncgo_prepare_output(out_xml, out_err) != 0) {
        return -1;
    }

    if (!datastore || !datastore[0]) {
        return ncgo_fail(out_err, "datastore is required");
    }

    if (!edit_content || !edit_content[0]) {
        return ncgo_fail(out_err, "edit_content is required");
    }

    struct nc_rpc *rpc = nc_rpc_editdata(
        datastore,
        NC_RPC_EDIT_DFLTOP_UNKNOWN,
        edit_content,
        NC_PARAMTYPE_CONST
    );
    return ncgo_dispatch_rpc(client, rpc, "nc_rpc_editdata failed", out_xml, out_err);
}

int ncgo_rpc_kill(
    ncgo_client_t *client,
    uint32_t session_id,
    char **out_xml,
    char **out_err
) {
    if (ncgo_prepare_output(out_xml, out_err) != 0) {
        return -1;
    }

    struct nc_rpc *rpc = nc_rpc_kill(session_id);
    return ncgo_dispatch_rpc(client, rpc, "nc_rpc_kill failed", out_xml, out_err);
}

void ncgo_close(ncgo_client_t *client) {
    if (!client) {
        return;
    }

    if (client->sess) {
        nc_session_free(client->sess, NULL);
        client->sess = NULL;
    }

    free(client->password_copy);
    free(client);
}

void ncgo_string_free(char *s) {
    free(s);
}

static int ncgo_dispatch_rpc(
    ncgo_client_t *client,
    struct nc_rpc *rpc,
    const char *build_context,
    char **out_xml,
    char **out_err
) {
    if (ncgo_prepare_output(out_xml, out_err) != 0) {
        nc_rpc_free(rpc);
        return -1;
    }

    if (!rpc) {
        return ncgo_fail_last(out_err, build_context);
    }

    return ncgo_exec_rpc(client, rpc, out_xml, out_err);
}

static int ncgo_exec_rpc(
    ncgo_client_t *client,
    struct nc_rpc *rpc,
    char **out_xml,
    char **out_err
) {
    int rc = -1;
    int print_rc = 0;
    struct lyd_node *envp = NULL;
    struct lyd_node *op = NULL;
    uint64_t msgid = 0;
    NC_MSG_TYPE msgtype;
    char *xml = NULL;

    if (!client || !client->sess || !rpc || !out_xml) {
        ncgo_fail(out_err, "invalid arguments");
        goto cleanup;
    }

    msgtype = nc_send_rpc(client->sess, rpc, NCGO_SEND_TIMEOUT_MS, &msgid);
    if (msgtype == NC_MSG_WOULDBLOCK) {
        ncgo_fail(out_err, "nc_send_rpc timed out");
        goto cleanup;
    }
    if (msgtype != NC_MSG_RPC) {
        ncgo_fail_last(out_err, "nc_send_rpc failed");
        goto cleanup;
    }

    for (;;) {
        msgtype = nc_recv_reply(client->sess, rpc, msgid, NCGO_REPLY_TIMEOUT_MS, &envp, &op);
        if (msgtype != NC_MSG_NOTIF) {
            break;
        }

        lyd_free_all(envp);
        lyd_free_all(op);
        envp = NULL;
        op = NULL;
    }

    if (msgtype == NC_MSG_WOULDBLOCK) {
        ncgo_fail(out_err, "nc_recv_reply timed out");
        goto cleanup;
    }
    if (msgtype == NC_MSG_ERROR) {
        ncgo_fail_last(out_err, "nc_recv_reply read failed");
        goto cleanup;
    }
    if (msgtype != NC_MSG_REPLY) {
        ncgo_fail(out_err, "nc_recv_reply returned unexpected message type");
        goto cleanup;
    }

    if (op) {
        print_rc = lyd_print_mem(&xml, op, LYD_XML, 0);
    } else if (envp) {
        print_rc = lyd_print_mem(&xml, envp, LYD_XML, 0);
    } else {
        ncgo_fail(out_err, "empty reply");
        goto cleanup;
    }

    if (print_rc || !xml) {
        ncgo_fail_last(out_err, "lyd_print_mem failed");
        goto cleanup;
    }

    *out_xml = xml;
    xml = NULL;
    rc = 0;

cleanup:
    free(xml);
    lyd_free_all(envp);
    lyd_free_all(op);
    nc_rpc_free(rpc);
    return rc;
}
