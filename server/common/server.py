import socket
import logging
import signal
from common.protocol import recv_batch, send_ack
from common.utils import store_bets


ACCEPT_TIMEOUT_SECONDS = 1


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._is_running = True
        self._server_socket.settimeout(ACCEPT_TIMEOUT_SECONDS)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        try:
            while self._is_running:
                try: 
                    client_sock = self.__accept_new_connection()
                    self.__handle_client_connection(client_sock)
                except socket.timeout:
                    continue
        except OSError as e:
            if self._is_running:
                logging.error(f"action: run | result: fail | error: {e}")
            else:
                pass 
        finally:
            self._server_socket.close()
            logging.info('action: close_resource | result: success | resource: server_socket')

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            while True:
                try:
                    bets = recv_batch(client_sock)
                except EOFError:
                    # Client closed the connection cleanly — all batches received
                    break
                store_bets(bets)
                logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')
                send_ack(client_sock, True)
        except (OSError, ValueError) as e:
            logging.error(f"action: apuesta_recibida | result: fail | error: {e}")
            try:
                send_ack(client_sock, False)
            except OSError:
                pass
        finally:
            client_sock.close()
            logging.info('action: close_resource | result: success | resource: client_socket')

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.debug('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c


    def _handle_signal(self, signum, frame):
        """
        Handle signal to graceful shutdown the server

        Set _is_running to False 
        """
        self._is_running = False
        logging.info('action: signal_received | result: success | signum: {}'.format(signum))

    
