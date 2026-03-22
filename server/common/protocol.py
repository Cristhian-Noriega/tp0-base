from common.utils import Bet


HEADER_SIZE = 2
BYTE_SHIFT = 8
BET_FIELD_COUNT = 6
ACK_SUCCESS = 0x01
ACK_FAILURE = 0x00

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
    
