export namespace config {
	
	export class SpotifyConfig {
	    client_id: string;
	    redirect_uri: string;
	
	    static createFrom(source: any = {}) {
	        return new SpotifyConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.client_id = source["client_id"];
	        this.redirect_uri = source["redirect_uri"];
	    }
	}
	export class Location {
	    lat: number;
	    lon: number;
	    city: string;
	
	    static createFrom(source: any = {}) {
	        return new Location(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lat = source["lat"];
	        this.lon = source["lon"];
	        this.city = source["city"];
	    }
	}
	export class Config {
	    location: Location;
	    weather_api_key: string;
	    spotify: SpotifyConfig;
	    polling_interval_idle_ms: number;
	    polling_interval_playing_ms: number;
	    min_session_duration_sec: number;
	    session_gap_sec: number;
	    first_run: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.location = this.convertValues(source["location"], Location);
	        this.weather_api_key = source["weather_api_key"];
	        this.spotify = this.convertValues(source["spotify"], SpotifyConfig);
	        this.polling_interval_idle_ms = source["polling_interval_idle_ms"];
	        this.polling_interval_playing_ms = source["polling_interval_playing_ms"];
	        this.min_session_duration_sec = source["min_session_duration_sec"];
	        this.session_gap_sec = source["session_gap_sec"];
	        this.first_run = source["first_run"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

export namespace diary {
	
	export class ArtistStats {
	    name: string;
	    total_minutes: number;
	    track_plays: number;
	    sessions: number;
	
	    static createFrom(source: any = {}) {
	        return new ArtistStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.total_minutes = source["total_minutes"];
	        this.track_plays = source["track_plays"];
	        this.sessions = source["sessions"];
	    }
	}
	export class DailyStats {
	    date_local: string;
	    total_listen_sec: number;
	    session_count: number;
	    track_count: number;
	    dominant_mood?: string;
	    avg_temp?: number;
	    dominant_weather?: string;
	
	    static createFrom(source: any = {}) {
	        return new DailyStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date_local = source["date_local"];
	        this.total_listen_sec = source["total_listen_sec"];
	        this.session_count = source["session_count"];
	        this.track_count = source["track_count"];
	        this.dominant_mood = source["dominant_mood"];
	        this.avg_temp = source["avg_temp"];
	        this.dominant_weather = source["dominant_weather"];
	    }
	}
	export class DayOfWeekStat {
	    day: number;
	    total_hours: number;
	    sessions: number;
	
	    static createFrom(source: any = {}) {
	        return new DayOfWeekStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.day = source["day"];
	        this.total_hours = source["total_hours"];
	        this.sessions = source["sessions"];
	    }
	}
	export class HourStat {
	    hour: number;
	    total_hours: number;
	    sessions: number;
	
	    static createFrom(source: any = {}) {
	        return new HourStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hour = source["hour"];
	        this.total_hours = source["total_hours"];
	        this.sessions = source["sessions"];
	    }
	}
	export class Insight {
	    icon: string;
	    title: string;
	    description: string;
	    stat: string;
	
	    static createFrom(source: any = {}) {
	        return new Insight(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.icon = source["icon"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.stat = source["stat"];
	    }
	}
	export class MoodStat {
	    mood: string;
	    count: number;
	    percentage: number;
	
	    static createFrom(source: any = {}) {
	        return new MoodStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mood = source["mood"];
	        this.count = source["count"];
	        this.percentage = source["percentage"];
	    }
	}
	export class Session {
	    id: string;
	    started_at: string;
	    ended_at?: string;
	    duration_sec: number;
	    track_count: number;
	    source: string;
	    mood?: string;
	    mood_set_at?: string;
	    weather_temp?: number;
	    weather_cond?: string;
	    weather_desc?: string;
	    weather_humid?: number;
	    weather_icon?: string;
	    day_of_week: number;
	    hour_of_day: number;
	    month: number;
	    year: number;
	    date_local: string;
	    time_of_day: string;
	
	    static createFrom(source: any = {}) {
	        return new Session(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.started_at = source["started_at"];
	        this.ended_at = source["ended_at"];
	        this.duration_sec = source["duration_sec"];
	        this.track_count = source["track_count"];
	        this.source = source["source"];
	        this.mood = source["mood"];
	        this.mood_set_at = source["mood_set_at"];
	        this.weather_temp = source["weather_temp"];
	        this.weather_cond = source["weather_cond"];
	        this.weather_desc = source["weather_desc"];
	        this.weather_humid = source["weather_humid"];
	        this.weather_icon = source["weather_icon"];
	        this.day_of_week = source["day_of_week"];
	        this.hour_of_day = source["hour_of_day"];
	        this.month = source["month"];
	        this.year = source["year"];
	        this.date_local = source["date_local"];
	        this.time_of_day = source["time_of_day"];
	    }
	}
	export class SessionTrack {
	    id: string;
	    session_id: string;
	    title: string;
	    artist: string;
	    album: string;
	    duration_ms: number;
	    listened_ms: number;
	    album_art_url?: string;
	    album_art_path?: string;
	    spotify_id?: string;
	    genres?: string;
	    energy?: number;
	    valence?: number;
	    tempo?: number;
	    source: string;
	    played_at: string;
	    position_in_session: number;
	
	    static createFrom(source: any = {}) {
	        return new SessionTrack(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.session_id = source["session_id"];
	        this.title = source["title"];
	        this.artist = source["artist"];
	        this.album = source["album"];
	        this.duration_ms = source["duration_ms"];
	        this.listened_ms = source["listened_ms"];
	        this.album_art_url = source["album_art_url"];
	        this.album_art_path = source["album_art_path"];
	        this.spotify_id = source["spotify_id"];
	        this.genres = source["genres"];
	        this.energy = source["energy"];
	        this.valence = source["valence"];
	        this.tempo = source["tempo"];
	        this.source = source["source"];
	        this.played_at = source["played_at"];
	        this.position_in_session = source["position_in_session"];
	    }
	}
	export class TrackStats {
	    title: string;
	    artist: string;
	    album: string;
	    play_count: number;
	    total_minutes: number;
	
	    static createFrom(source: any = {}) {
	        return new TrackStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.artist = source["artist"];
	        this.album = source["album"];
	        this.play_count = source["play_count"];
	        this.total_minutes = source["total_minutes"];
	    }
	}
	export class WeatherStat {
	    condition: string;
	    session_count: number;
	    total_hours: number;
	
	    static createFrom(source: any = {}) {
	        return new WeatherStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.condition = source["condition"];
	        this.session_count = source["session_count"];
	        this.total_hours = source["total_hours"];
	    }
	}

}

export namespace player {
	
	export class TrackInfo {
	    title: string;
	    artist: string;
	    album: string;
	    album_art_url?: string;
	    duration_ms: number;
	    position_ms: number;
	    genres?: string[];
	    source: string;
	    spotify_id?: string;
	    is_playing: boolean;
	    energy?: number;
	    valence?: number;
	    tempo?: number;
	    // Go type: time
	    detected_at: any;
	
	    static createFrom(source: any = {}) {
	        return new TrackInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.artist = source["artist"];
	        this.album = source["album"];
	        this.album_art_url = source["album_art_url"];
	        this.duration_ms = source["duration_ms"];
	        this.position_ms = source["position_ms"];
	        this.genres = source["genres"];
	        this.source = source["source"];
	        this.spotify_id = source["spotify_id"];
	        this.is_playing = source["is_playing"];
	        this.energy = source["energy"];
	        this.valence = source["valence"];
	        this.tempo = source["tempo"];
	        this.detected_at = this.convertValues(source["detected_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace weather {
	
	export class Data {
	    temp_celsius: number;
	    condition: string;
	    description: string;
	    humidity: number;
	    icon_code: string;
	    // Go type: time
	    fetched_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Data(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.temp_celsius = source["temp_celsius"];
	        this.condition = source["condition"];
	        this.description = source["description"];
	        this.humidity = source["humidity"];
	        this.icon_code = source["icon_code"];
	        this.fetched_at = this.convertValues(source["fetched_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

