echo "Change default Authenticator & Authorizer"
sed -ri 's/(authenticator:).*/\1 PasswordAuthenticator/' /etc/cassandra/cassandra.yaml
sed -ri 's/(authorizer:).*/\1 CassandraAuthorizer/' /etc/cassandra/cassandra.yaml
#test "$(hostname)" == 'cassandra-demo-dc1-rack2-0' && echo "update param" && sed -i 's/windows_timer_interval: 1/windows_timer_interval: 2/' /etc/cassandra/cassandra.yaml
#test "$(hostname)" == 'cassandra-demo-dc1-rack3-0' && echo "-Dcassandra.replace_address_first_boot=<pod_ip>" >> /etc/cassandra/jvm.options
echo "** end of pre_run.sh script, continue with run.sh"
