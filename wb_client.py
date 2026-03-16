import asyncio
import json
import websockets

async def main():
    uri = "ws://localhost:8080/resources"

    async with websockets.connect(uri) as ws:
        print("connected to", uri)

        for _ in range(5):  # 收 5 筆
            msg = await ws.recv()
            data = json.loads(msg)
            print(json.dumps(data, indent=2, default=str))

asyncio.run(main())