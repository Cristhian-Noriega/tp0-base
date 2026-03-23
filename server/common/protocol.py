from common.utils import Bet, load_bets, has_won


HEADER_SIZE = 2
BYTE_SHIFT = 8
BET_FIELD_COUNT = 6
ACK_SUCCESS = 0x01
ACK_FAILURE = 0x00
BYTE_MASK = 0xFF
NOT_READY = 0x00
READY = 0x01

def recv_exactly(conn, n) -> bytes:
    data = b''
    while len(data) < n:
        chunk = conn.recv(n - len(data))
        if not chunk:
            raise EOFError("Connection closed before receiving all data")
        data += chunk
    return data


def recv_bet(conn) -> Bet:
    raw_header = recv_exactly(conn, HEADER_SIZE)
    header = (raw_header[0] << BYTE_SHIFT) | raw_header[1]
    
    raw_payload = recv_exactly(conn, header)
    payload = raw_payload.decode('utf-8').splitlines()

    if len(payload) != BET_FIELD_COUNT:
        raise ValueError(f"Invalid bet payload. Expected {BET_FIELD_COUNT} fields, got {len(payload)}")

    return Bet(
        agency=payload[0],
        first_name=payload[1],
        last_name=payload[2],
        document=payload[3],
        birthdate=payload[4],
        number=payload[5]
    )
    

def recv_batch(conn) -> list:
    raw_count = recv_exactly(conn, HEADER_SIZE)
    count = (raw_count[0] << BYTE_SHIFT) | raw_count[1]

    bets = []
    for _ in range(count):
        bets.append(recv_bet(conn))
    return bets


def send_ack(conn, success: bool) -> None:
    conn.sendall(bytes([ACK_SUCCESS if success else ACK_FAILURE]))


def recv_winners_query(conn) -> int:
    """
    It receives the ID of the agency that wants to query the winners.
    [2 bytes: agency_id]
    """
    raw_agency_id = recv_exactly(conn, HEADER_SIZE)
    agency_id = (raw_agency_id[0] << BYTE_SHIFT) | raw_agency_id[1]
    return agency_id

def send_winners(conn, agency_id: int) -> None:
    """
    [2 bytes: cantidad de ganadores en bytes]
    [DNI\n DNI\n ...]
    """
    winners = [
        bet.document 
        for bet in load_bets() 
        if has_won(bet) and bet.agency == agency_id
    ]
    
    count = len(winners)
    header = bytes([count >> BYTE_SHIFT, count & BYTE_MASK])
    payload = "\n".join(winners) + "\n" if winners else ""
    
    conn.sendall(bytes([READY]) + header + payload.encode('utf-8'))

def send_not_ready(conn) -> None:
    """
    Sends to client not ready to get the winners.
    [1 byte to tell if it's ready or not]
    """
    conn.sendall(bytes([NOT_READY]))