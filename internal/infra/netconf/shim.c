#include "shim.h"

#include <stdlib.h>
#include <string.h>

#define MODULES_DIR "/home/nqs/nms_rc/third_party/libnetconf2/modules"


static int ncgo_fail(char **out_err, const char *msg) {
    if (out_err) {
        *out_err = strdup(msg ? msg : "unknown error");
    }
    return -1;
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
    printf("Connecting to %s:%u with username '%s'...\n", host, port, username);
    printf("Using schema search path: %s\n", schema_searchpath ? schema_searchpath : "(default)");
    if (!host || !username || !password || !out_client) {
        return ncgo_fail(out_err, "invalid arguments");
    }

    *out_client = NULL;

    ncgo_client_t *c = calloc(1, sizeof(*c));
    if (!c) {
        return ncgo_fail(out_err, "calloc failed");
    }

    c->password_copy = strdup(password);
    if (!c->password_copy) {
        free(c);
        return ncgo_fail(out_err, "strdup failed");
    }

    if (nc_client_set_schema_searchpath(MODULES_DIR)) {
        ncgo_close(c);
        return ncgo_fail(out_err, "nc_client_set_schema_searchpath failed");
    }

    nc_client_ssh_set_auth_pref(NC_SSH_AUTH_PASSWORD, 4);
    nc_client_ssh_set_username(username);
    nc_client_ssh_set_auth_password_clb(ncgo_password_cb, c->password_copy);

    nc_client_ssh_set_knownhosts_mode(NC_SSH_KNOWNHOSTS_ACCEPT_NEW);

    c->sess = nc_connect_ssh(host, port, NULL);
    if (!c->sess) {
        ncgo_close(c);
        return ncgo_fail(out_err, "nc_connect_ssh failed");
    }

    *out_client = c;
    return 0;
}

int ncgo_rpc(
    ncgo_client_t *client,
    NC_RPC_TYPE rpc_type,
    const char *filter,
    char **out_xml,
    char **out_err
) {
    int rc = -1;
    int r = 0;
    struct lyd_node *envp = NULL, *op = NULL;
    struct nc_rpc *rpc = NULL;
    uint64_t msgid = 0;
    NC_MSG_TYPE msgtype;
    char *xml = NULL;

    if (!client || !client->sess || !out_xml) {
        return ncgo_fail(out_err, "invalid arguments");
    }

    *out_xml = NULL;
    if (out_err) {
        *out_err = NULL;
    }

    switch (rpc_type) {
    case NC_RPC_GET:
        rpc = nc_rpc_get(filter, NC_WD_UNKNOWN, NC_PARAMTYPE_CONST);
        if (!rpc) {
            ncgo_fail(out_err, "nc_rpc_get failed");
            goto cleanup;
        }
        break;

    case NC_RPC_GETCONFIG:
        rpc = nc_rpc_getconfig(NC_DATASTORE_RUNNING, filter, NC_WD_UNKNOWN, NC_PARAMTYPE_CONST);
        if (!rpc) {
            ncgo_fail(out_err, "nc_rpc_getconfig failed");
            goto cleanup;
        }
        break;

    default:
        ncgo_fail(out_err, "unsupported rpc type");
        goto cleanup;
    }

    msgtype = nc_send_rpc(client->sess, rpc, 1000, &msgid);
    if (msgtype != NC_MSG_RPC) {
        ncgo_fail(out_err, "nc_send_rpc failed");
        goto cleanup;
    }

    msgtype = nc_recv_reply(client->sess, rpc, msgid, 10000, &envp, &op);
    if (msgtype != NC_MSG_REPLY) {
        ncgo_fail(out_err, "nc_recv_reply failed");
        goto cleanup;
    }

    if (op) {
        r = lyd_print_mem(&xml, op, LYD_XML, 0);
    } else if (envp) {
        r = lyd_print_mem(&xml, envp, LYD_XML, 0);
    } else {
        ncgo_fail(out_err, "empty reply");
        goto cleanup;
    }

    if (r || !xml) {
        ncgo_fail(out_err, "lyd_print_mem failed");
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