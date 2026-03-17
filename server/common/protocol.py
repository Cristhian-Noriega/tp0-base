from common.utils import Bet

def recv_exactly(conn, n) -> bytes:
    data = b''
    while len(data) < n:
        chunk = conn.recv(n - len(data))
        if not chunk:
            raise EOFError("Connection closed before receiving all data")
        data += chunk
    return data


def recv_bet(conn) -> Bet:
    raw_header = recv_exactly(conn, 2)
    header = (raw_header[0] << 8) | raw_header[1]
    
    raw_payload = recv_exactly(conn, header)
    payload = raw_payload.decode('utf-8').splitlines()

    if len(payload) != 6:
        raise ValueError(f"Invalid bet payload. Expected 6 fields, got {len(payload)}")

    return Bet(
        agency=payload[0],
        first_name=payload[1],
        last_name=payload[2],
        document=payload[3],
        birthdate=payload[4],
        number=payload[5]
    )
    

def send_ack(conn, success: bool) -> None:
    conn.sendall(bytes([0x01 if success else 0x00]))
    
