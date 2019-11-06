CQLSH_OUTPUT=$(/usr/bin/cqlsh -u'cassandra' -p'cassandra' -f /etc/liveness/command.cql 2>&1)
if [[ $? -eq 0 ]]; then
    if [[ $ CQLSH _OUTPUT == *" Cassandra 3. "* ]]; then
        ALIVE = true
    else
        CQLSH_OUTPUT/usr/bin/cqlsh -u'pns_cassandra_monitoring' -p'XXXXX' -f /etc/liveness/command.cql 2>&1
        if [[ $? -eq 0 ]]; then
            if [[ $ CQLSH _OUTPUT == *" Cassandra 3. "* ]]; then
                ALIVE = true
            else
                ALIVE =false
                fi
            else
                ALIVE =false
        fi
    fi
else
    CQLSH_OUTPUT=$(/usr/bin/cqlsh -u'pns_cassandra_monitoring' -p'XXXXX' -f /etc/liveness/command.cql 2>&1)
    if [[ $? -eq 0 ]]; then
        if [[ $ CQLSH _OUTPUT == *" Cassandra 3. "* ]]; then
            ALIVE = true
        else
            ALIVE =false
            fi
        else
            ALIVE =false
    fi
fi
