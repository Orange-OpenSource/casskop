FROM cassandra:3.11

RUN  apt-get update && apt-get -qq -y install \
       libcap2-bin \
     && setcap cap_ipc_lock=ep $(readlink -f $(which java))

#Make sure we don't use image entrypoint (the bootstrap is done by CassKop)
ENTRYPOINT []
#Just start Cassandra
CMD ["cassandra", "-f"]