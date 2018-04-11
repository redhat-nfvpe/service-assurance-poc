#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>
#include <limits.h>
#include <signal.h>
#include <sys/time.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <errno.h>

#include "proton/reactor.h"
#include "proton/message.h"
#include "proton/connection.h"
#include "proton/session.h"
#include "proton/link.h"
#include "proton/delivery.h"
#include "proton/event.h"
#include "proton/handlers.h"
#include "proton/transport.h"
#include "proton/url.h"

#include "proton-common.h"

#define GET_APP_DATA(handler) ((app_data_t *)pn_handler_mem(handler))

typedef struct {
  char *socket_path;
  int socket_fd;
  int credit;
  char *source;
  pn_message_t *message;
  //hosts_t host_addresses;
  const char *url;
  const char *host;
  const char *target;
  int prefetch;
  char *decode_buffer;
  size_t buffer_len;
  int debug;
  int received_msgs;
  //  time_t link_active;
} app_data_t;

static int done = 0;

#if 0
static time_t now()
{
  struct timeval tv;
  int rc = gettimeofday(&tv, NULL);
  if (rc) fatal("gettimeofday() failed");
  return (tv.tv_sec * 1000) + (time_t)(tv.tv_usec/1000);
}
#endif

static void process_message(app_data_t *app)
{
  //pn_atom_t id = pn_message_get_id(app->message);
  //const time_t then = pn_message_get_creation_time(app->message);
  //const uint64_t recv_seq = id.u.as_long;

  if (app->debug) {
    fprintf(stdout, "Message received!\n");
    pn_data_t *body  = pn_message_body(app->message);

    pn_data_next(body);
    pn_bytes_t b = pn_data_get_binary(body);
    printf("[%ld]:%s\n", b.size, b.start);

    //PN_BINARY
    if (app->socket_fd != 0) {
      if (write(app->socket_fd, b.start, b.size) < 0) {
	perror("write failed");
      }
    }
    //pn_free(b);

  }
}


static void event_handler(pn_handler_t *handler,
                          pn_event_t *event,
                          pn_event_type_t type)
{
  app_data_t *app = GET_APP_DATA(handler);

  switch(type) {
    
  case PN_CONNECTION_INIT:
    {
      // reactor is ready, create a link to the broaker
      pn_connection_t *conn;
      pn_session_t *session;
      pn_link_t *receiver;

      conn = pn_event_connection(event);
      pn_connection_open(conn);
      session = pn_session(conn);
      pn_session_open(session);
      receiver = pn_receiver(session,"amqpProxyReceiver");
      printf("target:%s\n", app->target);
      pn_terminus_set_address(pn_link_source(receiver),
			      app->target);
      pn_link_open(receiver);
      // cannot receive without granting credit:
      pn_link_flow(receiver, app->prefetch);
    }
    break;
  case PN_CONNECTION_UNBOUND:
    //printf("PN_CONNECTION_UNBOUND\n");
    pn_connection_release(pn_event_connection(event));
    break;

  case PN_LINK_REMOTE_OPEN:
    //app->link_active = now();
    break;

  case PN_LINK_REMOTE_CLOSE:
    // shutdown - clean up connection and session
    // this will cause the amain loop to eventually exit
    pn_session_close(pn_event_session(event));
    pn_connection_close(pn_event_connection(event));
    break;

  case PN_DELIVERY:
    {
      // A message has been received
      pn_delivery_t *dlv = pn_event_delivery(event);
      pn_link_t *receiver = pn_event_link(event);

      // A full message has arrived
      if (pn_delivery_readable(dlv) && !pn_delivery_partial(dlv)) {
	size_t len;
	// try to decode the message body
	if (pn_delivery_pending(dlv) > app->buffer_len) {
	  app->buffer_len = pn_delivery_pending(dlv);
	  free(app->decode_buffer);
	  app->decode_buffer = (char *)malloc(app->buffer_len);
	  if (!app->decode_buffer) {
	    fatal("cannot allocate buffer");
	  }
	}

	// read in the raw data
	len = pn_link_recv(receiver, app->decode_buffer, app->buffer_len);
	if (len > 0) {
	  // decode it into a proton message
	  pn_message_clear(app->message);
	  if (PN_OK == pn_message_decode(app->message,
					 app->decode_buffer,
					 len)) {
	    process_message(app);
	  }
	}

	if (!pn_delivery_settled(dlv)) {
	  // remote has not settled, so it is tracking the delivery.
	  // Ack it.
	  pn_delivery_update(dlv, PN_ACCEPTED);
	}

	// done with the delivery, move to the next add and free it
	pn_link_advance(receiver);
	pn_delivery_settle(dlv); // dlv is now freed
	app->received_msgs++;

	// replenish credit if it drops below 1/2 prefetch level
	int credit = pn_link_credit(receiver);
	if (credit < app->prefetch / 2) {
	  pn_link_flow(receiver, app->prefetch - credit);
	}
      }
    } // case PN_DELIVERY
    break;
  case PN_TRANSPORT_ERROR:
    {
      // the connection to the peer failed.
      pn_transport_t *tport = pn_event_transport(event);
      pn_condition_t *cond = pn_transport_condition(tport);
      fprintf(stderr, "Network transport failed!\n");
      if (pn_condition_is_set(cond)) {
	const char *name = pn_condition_get_name(cond);
	const char *desc = pn_condition_get_description(cond);
	fprintf (stderr, "    Error: %s  Description: %s\n",
		 (name) ? name : "<error name not provided>",
		 (desc) ? desc : "<no description provided>");
	// pn_reactor_process will exit with a false return value, stopping
	// the main loop
      }
    }
    break;

  default:
    // ignore the rest
    break;
  } // switch

}

static void delete_handler(pn_handler_t *handler)
{
  app_data_t *app = GET_APP_DATA(handler);

  if (app->message) {
    pn_decref(app->message);
    app->message = NULL;
  }

  free(app->decode_buffer);
}

static void stop(int sig)
{
  done = 1;
}

static void usage(char *argv[])
{
  printf("Usage: %s <options>\n", argv[0]);
  //printf("-a      \tThe host address [localhost:5672]\n");
  //printf("-t      \tTopic name [ReceiveExample]\n");
  printf("-u      \tURL [amqp://localhost:5672/collectd]\n");
  printf("-p      \tPre-fetch window size [100]\n");
  printf("-s      \tUnix socket to connect\n");
  exit(1);
}

static int init_app_data (int argc, char *argv[], app_data_t *app)
{
  memset(app, 0, sizeof(app_data_t));

  app->debug = 1;
  app->socket_fd = 0;
  app->url = strdup("amqp://127.0.0.1:5672/collectd/telemetry");
  app->buffer_len = 64;
  app->decode_buffer = malloc(app->buffer_len);
  if (!app->decode_buffer) {
    fatal("Cannot allocate decode buffer");
  }
  app->prefetch = 100;
  app-> message = pn_message();
  if (!app->message) {
    fatal("Message allocation failed");
  }

  opterr = 0;
  int c;
  // option:
  // -h : help
  // -u <str>: url
  // -p <int>: prefetch
  // -s <socket> : unix socket

  while ((c= getopt(argc, argv, "a:t:p:hu:s:")) != -1) {
    switch(c) {
    case 'u':
      app->url = optarg;
      break;
    case 'h':
      usage(argv);
      break;
    case 'p':
      app->prefetch = atoi(optarg);
      break;
    case 's':
      app->socket_path = optarg;
      break;
    }
  }

  if (app->prefetch <= 0) {
    fatal("prefetch must be >= zero");
  }

  return 0;
}

// create a connection to the server
static void amqp1_connect (app_data_t *data,
			   pn_reactor_t *reactor,
			   pn_handler_t *handler)
{
  pn_connection_t *conn = NULL;
  //const char *host = hosts_get(&data->host_addresses);
  pn_url_t *url = pn_url_parse(data->url);
  const char *host = pn_url_get_host(url);
  data->target = strdup(pn_url_get_path(url));

  if (data->debug) {
    //XXX
    printf("host: %s\n", host);
    printf("port: %s\n", pn_url_get_port(url));
    printf("target: %s\n", data->target);
  }

  if (url == NULL) {
    fprintf (stderr, "Invalid host address %s\n", host);
    exit (1);
  }

  if (data->debug) {
    fprintf(stdout, "Connecting to %s...\n", data->url);
  }
  conn = pn_reactor_connection_to_host (reactor,
                                        pn_url_get_host (url),
                                        pn_url_get_port (url),
                                        handler);
  pn_decref(url);
  // the container name should be unique for each client
  // attached to the broker
  {
    char hname[HOST_NAME_MAX + 1] = "<unknown host>";
    char cname[256];

    gethostname (hname, sizeof (hname));
    snprintf (cname, sizeof (cname), "amqp-proxy-container-%s-%d-%d",
              hname, getpid (), rand ());

    pn_connection_set_container (conn, cname);
  }
}

int main(int argc, char *argv[])
{
  errno = 0;
  signal(SIGINT, stop);
  signal(SIGALRM, stop);

  pn_handler_t *handler = pn_handler_new(event_handler,
					 sizeof(app_data_t),
					 delete_handler);
  app_data_t *app = GET_APP_DATA(handler);
  if (init_app_data(argc, argv, app) != 0) {
    exit(1);
  }

  if (app->socket_path != NULL) {
    /* connect socket*/
    app->socket_fd = socket(AF_UNIX, SOCK_STREAM, 0);
    if (app->socket_fd < 0) {
      perror("socket error");
      exit(1);
    }
    struct sockaddr_un addr;
    memset(&addr, 0, sizeof(struct sockaddr_un));
    addr.sun_family = AF_UNIX;
    strcpy(addr.sun_path, app->socket_path);

    if(connect(app->socket_fd,
	       (struct sockaddr *)&addr, sizeof(struct sockaddr_un)) < 0) {
      perror("connect error");
      exit(-1);
    }
  } else {
	  fprintf(stderr, "no unix path!\n");
	  usage(argv);
	  exit(-1);
  }

  /* Attach the pn_handshaker() handler.  This handler deals with endpoint
   * events from the peer so we don't have to.
   */
  pn_handler_add(handler, pn_handshaker());

  pn_reactor_t *reactor = pn_reactor();

  // make pn_reactor_process() wakeup every second
  pn_reactor_set_timeout(reactor, 1000);
  pn_reactor_start(reactor);

  amqp1_connect(app, reactor, handler);

  while (!done && pn_reactor_process(reactor)) {
  }

  return 0;
}
