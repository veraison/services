## Configuration

- `server-addr` (optional): address of the VTS server in the form
  `<host>:<port>`. If not specified, this defaults to `127.0.0.1:50051`.
  Unless `listen-addr` is specified (see below), VTS server will extract the
  port to listen on from this setting (but will listen on all local interfaces)
- `listen-addr` (optional): The address the VTS server will listen on in the
  form `<host>:<port>`. Only specify this if you want to restrict the server to
  listen on a particular interface; otherwise, the server will listen on all
  interfaces on the port specified in `server-addr`.

The special address format `vsock://<cid>:<port>` can be used for VM sockets,
e.g., when VTS runs in a separate VM and exchanges gRPC messages with the
frontend services through the
[virtio-vsock](https://vmsplice.net/~stefan/stefanha-kvm-forum-2015.pdf) device.
The `<cid>` value is ignored on the server side.
