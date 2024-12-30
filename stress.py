import asyncio
import aiohttp
from collections import Counter
import time
import timeit

URL = "http://localhost:33327/api/" 
URL2 = "http://localhost:33327/api/ping"  # Replace with your server URL
response_counter = Counter()


async def send_request(session,URL):
    try:
        
        async with session.get(URL) as response:
            response_text = await response.text()
            response_counter[response_text] += 1
        
        
    except aiohttp.ClientError as e:
        print(f"Request failed: {e}")

async def run_requests(rp,url):
    global response_counter
    response_counter = Counter()
    async with aiohttp.ClientSession() as session:
        tasks = [send_request(session,url) for _ in range(rp)]
        await asyncio.gather(*tasks)

    print("Response counts:",end="")
    for response_text, count in response_counter.items():
        print(f"{count}")

def main(url):
    rp = 10000
    asyncio.run(run_requests(rp,url))

def run1():
    main(URL)


def run2():
    main(URL2)

if __name__ == "__main__":
    execution_time = timeit.timeit(run2,number=1)
    print(f"Execution time: {execution_time} seconds")
    # execution_time2 = timeit.timeit(run2,number=1)
    # print(f"Execution time: {execution_time2} seconds")

   