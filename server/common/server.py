import socket
import logging
import signal
from common.protocol import recv_batch, send_ack, recv_winners_query, send_winners
from common.utils import store_bets
import threading


ACCEPT_TIMEOUT_SECONDS = 1


class Server:
    def __init__(self, port, listen_backlog, agencies_amount):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._is_running = True
        self._server_socket.settimeout(ACCEPT_TIMEOUT_SECONDS)
        self._agencies_amount = agencies_amount
        self._closed_connections = 0
        self._new_opened_connections = 0
        self._lock = threading.Lock()
        self._barrier = threading.Barrier(agencies_amount)

    def run(self):
        """
        Server loop that accepts new connections and spawns a thread for each.
        """
        logging.info("action: server_run | result: success")
        try:
            while self._is_running:
                try: 
                    client_sock = self.__accept_new_connection()
                    thread = threading.Thread(target=self.__handle_client_connection, args=(client_sock,))
                    thread.start()
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
                    if not bets: 
                        # Agency finished sending bets
                        break
                except EOFError:
                    break
                with self._lock:
                    store_bets(bets)
                logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')
                send_ack(client_sock, True)
            
            # Phase 2: Wait for all agencies to finish betting
            logging.info(f"action: wait_barrier | result: in_progress")
            arrival = self._barrier.wait()
            
            # Only one thread should log the sorteo success
            if arrival == 0:
                logging.info("action: sorteo | result: success")

            # Phase 3: Send winners
            self.__handle_winners_query(client_sock)

        except (OSError, ValueError) as e:
            logging.error(f"action: apuesta_recibida | result: fail | error: {e}")
        except threading.BrokenBarrierError:
            logging.error("action: wait_barrier | result: fail | error: barrier_broken")
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
        c.settimeout(ACCEPT_TIMEOUT_SECONDS) 
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c


    def _handle_signal(self, signum, frame):
        """
        Handle signal to graceful shutdown the server

        Set _is_running to False 
        """
        self._is_running = False
        logging.info('action: signal_received | result: success | signum: {}'.format(signum))

    def __handle_winners_query(self, client_sock):
        """
        Read message from a specific client and close the socket
 
        It is used after the first 5 connections are closed and to handle the winners query
        """
        try:
            agency_id = recv_winners_query(client_sock)
            send_winners(client_sock, agency_id)
            with self._lock:
                self._new_opened_connections += 1
        except (OSError, ValueError) as e:
            logging.error(f"action: winners_query | result: fail | error: {e}")
