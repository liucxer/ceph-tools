package ceph_conf_test

import (
	"ceph-tools/pkg/ceph_conf"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

var cephConfData = `[global]
fsid = 9ce3c024-c7d8-41e2-a884-c217f4c39465
mon_initial_members = node28
mon_host = 10.0.20.28
auth_cluster_required = none
auth_service_required = none
auth_client_required = none
osd_crush_update_on_start = false
osd_pool_default_size = 3
osd_pool_default_min_size = 1
mon_health_preluminous_compat_warning = false
mon_allow_pool_delete = true
ms_bind_port_min = 6800
ms_bind_port_max = 7800
infinity_mds_audit_delete = false
osd_failsafe_full_ratio = 1.0
mon_osd_full_ratio = 0.95
mon_osd_backfillfull_ratio = 0.95
mon_osd_nearfull_ratio = 0.80
mon_max_pg_per_osd = 800
osd_scrub_begin_hour = 0
osd_scrub_end_hour = 7
mon_pg_warn_max_object_skew = 20
osd_agent_max_ops = 8
osd_agent_max_low_ops = 4
osd_pool_erasure_code_stripe_unit = 65536
osd_heartbeat_interval = 12
osd_heartbeat_grace = 30
mon_osd_min_down_reporters = 2
mon_osd_min_in_ratio = 0.65
mon_osd_down_out_interval = 300
mon_osd_max_creating_pgs = 8192
ms_async_max_op_threads = 8
ms_async_op_threads = 5
rgw_thread_pool_size = 200
rgw_max_put_size = 10737418240
rgw_obj_stripe_size = 4194304
rgw_put_obj_min_window_size = 33554432
rgw_get_obj_window_size = 33554432
rgw_num_rados_handles = 8
rgw_ops_log_socket_path = /var/run/ceph/unix_socket
rgw_log_http_headers = "http_x_amz_copy_source"
rgw_enable_usage_log = true
rgw_enable_ops_log = true
rbd_cache = false
rbd_cache_writethrough_until_flush = false
bluestore_cache_autotune = true
bluestore_cache_kv_ratio = 0.2
bluestore_cache_meta_ratio = 0.8
bluestore_rocksdb_options = "compression=kNoCompression,max_write_buffer_number=32,min_write_buffer_number_to_merge=2,recycle_log_file_num=4,write_buffer_size=536870912,writable_file_max_buffer_size=0,compaction_readahead_size=2097152"
osd_max_backfills = 10
osd_recovery_max_active = 10
osd_recovery_max_single_start = 5
osd_recovery_sleep = 0.15
ms_bind_ipv4 = true
ms_bind_ipv6 = false
public_network = 10.0.20.0/24

[mon]
mon_clock_drift_allowed = 1.0
mon_pg_min_inactive = 0
mon_data_avail_warn = 10
mon_lease = 10
mon_warn_on_pool_pg_num_not_power_of_two = false

[mds]
mds_reconnect_timeout = 10
mds_bal_fragment_size_max = 200000
mds_max_purge_files = 2048
mds_max_purge_ops_per_pg = 2
mds_session_blacklist_on_evict = false
mds_session_blacklist_on_timeout = false

[osd]
osd data = /var/lib/ceph/osd/$cluster-$id
bluestore = true
osd_max_pg_per_osd_hard_ratio = 3.0
osd_memory_target = 2147483648
osd_op_num_shards=1
osd_delete_sleep_hdd=0
osd_delete_sleep_hybrid=0
osd_op_queue_mclock_client_op_lim= 99999
osd_op_queue_mclock_client_op_res= 1000.000000
osd_op_queue_mclock_client_op_wgt= 500.000000
osd_op_queue_mclock_osd_rep_op_lim= 99999
osd_op_queue_mclock_osd_rep_op_res= 1000.000000
osd_op_queue_mclock_osd_rep_op_wgt= 500.000000
osd_op_queue_mclock_peering_event_lim= 0.001000
osd_op_queue_mclock_peering_event_res= 0.000000
osd_op_queue_mclock_peering_event_wgt= 1.000000
osd_op_queue_mclock_pg_delete_lim= 99999.000000
osd_op_queue_mclock_pg_delete_res= 1000.000000
osd_op_queue_mclock_pg_delete_wgt= 500.000000
osd_op_queue_mclock_recov_lim= 0.001000
osd_op_queue_mclock_recov_res= 0.000000
osd_op_queue_mclock_recov_wgt= 1.000000
osd_op_queue_mclock_scrub_lim= 0.001000
osd_op_queue_mclock_scrub_res= 0.000000
osd_op_queue_mclock_scrub_wgt= 1.000000
osd_op_queue_mclock_snap_lim= 0.001000
osd_op_queue_mclock_snap_res= 0.000000
osd_op_queue_mclock_snap_wgt= 1.000000
osd_recovery_max_active= 1000
osd_recovery_max_single_start= 1
osd_max_backfills= 1000
osd_op_queue=mclock_opclass
osd_sa_schdule_interval=0

[osd.5]
debug_none=2/5
#debug_osd=2/5

[osd.6]
debug_none=2/5
#debug_osd=2/5

[osd.7]
debug_none=2/5
#debug_osd=2/5

[osd.8]
debug_none=2/5
#debug_osd=2/5

[client.admin]
admin_socket = /var/run/ceph/$cluster-$type.$id.$pid.$cctid.asok`

func TestCephConf_UnmarshalJSON(t *testing.T) {
	var cephConf ceph_conf.CephConf

	err := cephConf.UnmarshalJSON([]byte(cephConfData))
	require.NoError(t, err)

	err = json.Unmarshal([]byte(cephConfData), &cephConf)
	require.NoError(t, err)
}
