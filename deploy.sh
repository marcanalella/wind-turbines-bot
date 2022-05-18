rm -rf turbine-northern-power-bot

git clone git@github.com:marcanalella/turbine-northern-power-bot.git

docker build -t bot ./turbine-northern-power-bot

docker stop bot-turbine || true && docker rm bot-turbine || true

docker run -d --name bot-turbine -p 5005:5000 bot
