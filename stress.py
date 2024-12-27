import asyncio
import aiohttp
from collections import Counter
import time

URL = "http://localhost:8080/api/ping"  # Replace with your server URL
response_counter = Counter()


async def send_request(session):
    try:
        
        async with session.get(URL) as response:
            response_text = await response.text()
            response_counter[response_text] += 1
        
        
    except aiohttp.ClientError as e:
        print(f"Request failed: {e}")

async def main(rp):
    async with aiohttp.ClientSession() as session:
        tasks = [send_request(session) for _ in range(rp)]
        await asyncio.gather(*tasks)
    
    print("Response counts:")
    for response_text, count in response_counter.items():
        print(f"{response_text}: {count}")


if __name__ == "__main__":
    rp = 20000
    start = time.time()
    asyncio.run(main(rp))
    print(time.time()-start)