import predict_events
import ratings
import storage


def main():
    conn = storage.connect()
    try:
        print("ratings updated for events:", ratings.apply_resolved_events(conn))
        print("predictions:", predict_events.predict_open_events(conn))
    finally:
        conn.close()


if __name__ == "__main__":
    main()