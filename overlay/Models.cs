using System.Collections.Generic;

namespace LolBuildOverlay
{
    public class ItemInfo
    {
        public string id { get; set; }
        public string name { get; set; }
        public int gold { get; set; }
    }

    public class ItemChange
    {
        public string from;
        public string to;
        public string why;
    }

    public class RecommendedBuild
    {
        public string[] Items = new string[5];
        public string Boots;
        public ItemChange[] Changes = new ItemChange[0];
        public int NextIndex;
    }

    public class LiveItem
    {
        public int itemID { get; set; }
        public int slot { get; set; }
        public int count { get; set; }
        public string displayName { get; set; }
        public int price { get; set; }
    }

    public class LiveScores
    {
        public int assists { get; set; }
        public int creepScore { get; set; }
        public int deaths { get; set; }
        public int kills { get; set; }
        public float wardScore { get; set; }
    }

    public class LivePlayer
    {
        public string championName { get; set; }
        public bool isBot { get; set; }
        public List<LiveItem> items { get; set; }
        public int level { get; set; }
        public string position { get; set; }
        public string role { get; set; }
        public LiveScores scores { get; set; }
        public string summonerName { get; set; }
        public string riotIdGameName { get; set; }
        public string riotId { get; set; }
        public string team { get; set; }
    }

    public class ActivePlayer
    {
        public string summonerName { get; set; }
        public string riotIdGameName { get; set; }
        public string riotId { get; set; }
        public double currentGold { get; set; }
        public int level { get; set; }
    }

    public class LiveGameInfo
    {
        public string gameMode { get; set; }
        public double gameTime { get; set; }
        public string mapName { get; set; }
    }

    public class LiveGameData
    {
        public ActivePlayer activePlayer { get; set; }
        public List<LivePlayer> allPlayers { get; set; }
        public LiveGameInfo gameData { get; set; }
    }

    public class MatchState
    {
        public LivePlayer Me;
        public List<LivePlayer> Allies = new List<LivePlayer>();
        public List<LivePlayer> Enemies = new List<LivePlayer>();
        public double GameTime;
        public double Gold;
        public bool InGame;
    }
}
