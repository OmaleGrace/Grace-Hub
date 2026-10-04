import sys

import storage


def main():
    if len(sys.argv) < 2:
        sys.exit("usage: python main.py <event id>")
    event_id = int(sys.argv[1])

    conn = storage.connect()
    try:
        first = storage.save_prediction(conn, event_id, "home", 0.62, "py-test-v1")
        print("first save:", first)
        second = storage.save_prediction(conn, event_id, "away", 0.90, "py-test-v1")
        print("second save (same model version):", second)
    finally:
        conn.close()


if __name__ == "__main__":
    main()