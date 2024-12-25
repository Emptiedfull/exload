import asyncio
import aiohttp
from collections import Counter
import time

URL = "http://localhost:8080/api/ping"  # Replace with your server URL
response_counter = Counter()
response_times = []

async def send_request(session):
    try:
        start_time = time.time()
        async with session.get(URL) as response:
            response_text = await response.text()
            response_counter[response_text] += 1
        end_time = time.time()
        response_times.append(end_time - start_time)
    except aiohttp.ClientError as e:
        print(f"Request failed: {e}")

async def main(rp):
    async with aiohttp.ClientSession() as session:
        tasks = [send_request(session) for _ in range(rp)]
        await asyncio.gather(*tasks)
    
    print("Response counts:")
    for response_text, count in response_counter.items():
        print(f"{response_text}: {count}")

    if response_times:
        average_response_time = sum(response_times) / len(response_times)
        print(f"Average response time: {average_response_time:.2f} seconds")

if __name__ == "__main__":
    rp = 1000
    asyncio.run(main(rp))