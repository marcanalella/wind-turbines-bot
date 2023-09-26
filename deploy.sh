rm -rf wind-turbines-bot

git clone git@github.com:marcanalella/wind-turbines-bot.git

docker build -t bot ./wind-turbines-bot

docker stop bot-turbine || true && docker rm bot-turbine || true

docker run -d --name bot-turbine -p 5005:5000 bot
