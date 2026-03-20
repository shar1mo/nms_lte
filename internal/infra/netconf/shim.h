#ifndef NMS_NETCONF_SHIM_H
#define NMS_NETCONF_SHIM_H

#include <stdint.h>

#include <libyang/libyang.h>
#include <nc_client.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct ncgo_client {
    struct nc_session *sess;
    char *password_copy;
} ncgo_client_t;

int ncgo_client_init(void);
void ncgo_client_destroy(void);

int ncgo_connect_ssh(
    const char *host,
    uint16_t port,
    const char *username,
    const char *password,
    const char *schema_searchpath,
    ncgo_client_t **out_client,
    char **out_err
);

int ncgo_rpc(
    ncgo_client_t *client,
    NC_RPC_TYPE rpc_type,
    const char *payload,
    char **out_xml,
    char **out_err
);

int ncgo_rpc_get(
    ncgo_client_t *client,
    const char *filter,
    char **out_xml,
    char **out_err
);

int ncgo_rpc_getconfig(
    ncgo_client_t *client,
    const char *datastore,
    const char *filter,
    char **out_xml,
    char **out_err
);

int ncgo_rpc_edit(
    ncgo_client_t *client,
    const char *datastore,
    const char *edit_content,
    char **out_xml,
    char **out_err
);

int ncgo_rpc_copy(
    ncgo_client_t *client,
    const char *target,
    const char *url_trg,
    const char *source,
    const char *url_or_config_src,
    char **out_xml,
    char **out_err
);

int ncgo_rpc_delete(
    ncgo_client_t *client,
    const char *target,
    const char *url,
    char **out_xml,
    char **out_err
);

int ncgo_rpc_lock(
    ncgo_client_t *client,
    const char *datastore,
    char **out_xml,
    char **out_err
);

int ncgo_rpc_unlock(
    ncgo_client_t *client,
    const char *datastore,
    char **out_xml,
    char **out_err
);

int ncgo_rpc_commit(
    ncgo_client_t *client,
    char **out_xml,
    char **out_err
);

int ncgo_rpc_discard(
    ncgo_client_t *client,
    char **out_xml,
    char **out_err
);

int ncgo_rpc_cancel(
    ncgo_client_t *client,
    const char *persist_id,
    char **out_xml,
    char **out_err
);

int ncgo_rpc_validate(
    ncgo_client_t *client,
    const char *source,
    const char *url_or_config,
    char **out_xml,
    char **out_err
);

int ncgo_rpc_getschema(
    ncgo_client_t *client,
    const char *identifier,
    const char *version,
    const char *format,
    char **out_xml,
    char **out_err
);

int ncgo_rpc_subscribe(
    ncgo_client_t *client,
    const char *stream_name,
    const char *filter,
    const char *start_time,
    const char *stop_time,
    char **out_xml,
    char **out_err
);

int ncgo_rpc_getdata(
    ncgo_client_t *client,
    const char *datastore,
    const char *filter,
    char **out_xml,
    char **out_err
);

int ncgo_rpc_editdata(
    ncgo_client_t *client,
    const char *datastore,
    const char *edit_content,
    char **out_xml,
    char **out_err
);

int ncgo_rpc_kill(
    ncgo_client_t *client,
    uint32_t session_id,
    char **out_xml,
    char **out_err
);

void ncgo_close(ncgo_client_t *client);
void ncgo_string_free(char *s);

#ifdef __cplusplus
}
#endif

#endif