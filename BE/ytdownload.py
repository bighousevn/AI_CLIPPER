from pytubefix import YouTube
from pytubefix.cli import on_progress

url = 'https://www.youtube.com/watch?v=SOG0GmKts_I'
yt = YouTube(url)
print(f'Title: {yt.title}')

ys = yt.streams.get_highest_resolution()

ys.download()