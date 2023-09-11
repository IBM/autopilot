import asyncio
import random
import aiohttp


# ANSI colors
c = (
    "\033[0m",   # End of color
    "\033[36m",  # Cyan
    "\033[91m",  # Red
    "\033[35m",  # Magenta
)

async def makeconnection(idx: int):
    url='http://10.128.2.60:3333/status?host=epoc1work1&check=dcgm&r=1'
    print(c[idx + 1] + f"Initiated makeconnection({idx}).")
    async with aiohttp.ClientSession() as session:
        async with session.get(url) as resp:
            reply = await resp.text()
    return reply


async def main():
    #res = await asyncio.gather(*(makerandom(i, 10 - i - 1) for i in range(3)))
    res = await asyncio.gather(*(makeconnection(i) for i in range(3)))
    return res

if __name__ == "__main__":
    random.seed(444)
    r1, r2, r3 = asyncio.run(main())
    print()
    print(f"r1: {r1}, r2: {r2}, r3: {r3}")