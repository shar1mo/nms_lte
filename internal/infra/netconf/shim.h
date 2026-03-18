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
    const char *filter,
    char **out_xml,
    char **out_err
);

void ncgo_close(ncgo_client_t *client);
void ncgo_string_free(char *s);

#ifdef __cplusplus
}
#endif

#endif