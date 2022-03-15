WSNet2 Lobby API
================

## Create Room

POST /rooms

### リクエスト

### 成功レスポンス

### エラーレスポンス
| 概要 | HTTP Status (ResponseType) | gRPC Code | 発生箇所  | 備考 |
|------|----------------------------|-----------|-----------|------|
| レスポンスのmsgpackエンコード失敗 | InternalServerError | - | lobby/service/api.go: renderResponse() | - |
| ユーザ認証失敗 | Unauthorized | - | lobby/service/api.go: LobbyService.authUser() | - |
| リクエストbodyのmsgpackデコード失敗 | BadRequest | - | lobby/service/api.go: handleCreateRoom() | - |
| appIdのAppが無い | InternalServerError | - | lobby/room.go: RoomService.Create() | ユーザ認証失敗しているはずなので起こらない |
| gameサーバ取得失敗 | InternalServerError | - | lobby/game_cache.go: GameCache.Rand() | 生きているgameが見つからない |
| gRPC ClientをPoolから取得失敗 | InternalServerError | - | lobby/room.go: RoomService.Create() | - |
| gRPCタイムアウト | InternalServerError | DeadlineExceeded | lobby/room.go: RoomService.Create() | lobby側で設定したタイムアウト |
| appIdのAppが無い | InternalServerError | Internal | game/service/grpc.go: GameService.Create() | ユーザ認証失敗しているはずなので起こらない |
| context timeout | InternalServerError | DeadlineExceeded | game/repository.go: Repository.CreateRoom(), game/room.go: NewRoom() | game側で設定したタイムアウト |
| roomレコード作成失敗 | InternalServerError | Internal | game/repository.go: Repository.newRoomInfo() | - |
| room数上限 | **200 OK** (RoomLimit) | ResourceExhausted | game/repository.go: Repository.CreateRoom() | - |
| DB Commit失敗 | InternalServerError | Internal | game/repository.go: Repository.CreateRoom() | - |
| {public,private}PropsのUnmarshal失敗| BadRequest | InvalidArgument | game/room.go: NewRoom() | - |
| Player PropsのUnmarshal失敗 | BadRequest | InvalidArgument | game/client.go: newClient() | - |


## Join Room

POST /rooms/join/id/{roomId}
POST /rooms/join/number/{roomNumber}

### エラーレスポンス
| 概要 | HTTP Status (ResponseType) | gRPC Code | 発生箇所  | 備考 |
|------|----------------------------|-----------|-----------|------|
| レスポンスのmsgpackエンコード失敗 | InternalServerError | - | lobby/service/api.go: renderResponse() | - |
| ユーザ認証失敗 | Unauthorized | - | lobby/service/api.go: LobbyService.authUser() | - |
| リクエストbodyのmsgpackデコード失敗 | BadRequest | - | lobby/service/api.go: handleCreateRoom() | - |
| RoomIDが空 | BadRequest | - | lobby/service/api.go: handleJoinRoom() | - |
| RoomNumberが空または0 | BadRequest | - | lobby/service/api.go: handleJoinRoomByNumber() | - |
| appIdのAppが無い | InternalServerError | - | lobby/room.go: RoomService.JoinBy{Id,Number}() | ユーザ認証失敗しているはずなので起こらない |
| 入室可能なRoomが見つからない | **200 OK** (NoRoomFound) | - | lobby/room.go: RoomService.JoinBy{Id,Number}() | - |
| プロパティクエリ条件に合致しない | **200 OK** (NoRoomFound) | - | lobby/room.go: RoomService.JoinBy{Id,Number}() | - |
| publicPropsのデコード失敗 | InternalServerError | - | obby/room.go: RoomService.JoinBy{Id,Number}() | - |
| gameサーバ取得失敗 | InternalServerError | - | lobby/game_cache.go: GameCache.Get() | - |
| gRPC ClientをPoolから取得失敗 | InternalServerError | - | lobby/room.go: RoomService.join() | - |
| gRPCタイムアウト | InternalServerError | DeadlineExceeded | lobby/room.go: RoomService.join() | lobby側で設定したタイムアウト |
| appIdのAppが無い | InternalServerError | Internal | game/service/grpc.go: GameService.Join() | ユーザ認証失敗しているはずなので起こらない |
| gRPCタイムアウト | InternalServerError | Deadlineexceeded | game/repository.go: Repository.joinRoom() | game側で設定したタイムアウト |
| Roomが既に消えた | **200 OK** (NoRoomFound) | NotFound | game/repository.go: Repository.joinRoom() | lobbyでのチェック後に消えたパターン |
| Joinableでない | **200 OK** (NoRoomFound) | FailedPrecondition | game/room.go: msgJoin() | lobbyでのチェック後に折られた |
| 既に入室済み | Conflict | AlreadyExists | game/room.go: msgJoin() | Watcherとして既存も含む |
| 満室 | **200 OK** (RoomFull) | ResourceExhausted | game/room.go: msgJoin() | - |
| Player PropsのUnmarshal失敗 | BadRequest | InvalidArgument | game/client.go: newClient() | - |


## Random Join

POST /rooms/join/random/{searchGroup}

### エラーレスポンス
| 概要 | HTTP Status (ResponseType) | gRPC Code | 発生箇所  | 備考 |
|------|-------------|-----------|-----------|------|
| レスポンスのmsgpackエンコード失敗 | InternalServerError | - | lobby/service/api.go: renderResponse() | - |
| ユーザ認証失敗 | Unauthorized | - | lobby/service/api.go: LobbyService.authUser() | - |
| リクエストbodyのmsgpackデコード失敗 | BadRequest | - | lobby/service/api.go: handleJoinAtRandom() | - |
| タイムアウト | InternalServerError | - | lobby/room.go: RoomService.JoinAtRandom() | lobby側で設定したタイムアウト |
| GameCacheからの取得失敗 | InternalServerError | - | lobby/room_cache.go: roomCacheQuery.do() | - |
| Player PropsのUnmarshal失敗 | BadRequest | InvalidArgument | game/client.go: newClient() | - |
| 入室可能な部屋が見つからない | **200 OK** (NoRoomFound) | - | lobby/room.go: JoinAtRandom() | - |

※InvalidArgument以外のgRPCエラーは無視し別の部屋への入室を試行します


## Search Rooms

POST /rooms/search

### エラーレスポンス
| 概要 | HTTP Status (ResponseType) | gRPC Code | 発生箇所  | 備考 |
|------|----------------------------|-----------|-----------|------|
| レスポンスのmsgpackエンコード失敗 | InternalServerError | - | lobby/service/api.go: renderResponse() | - |
| ユーザ認証失敗 | Unauthorized | - | lobby/service/api.go: LobbyService.authUser() | - |
| リクエストbodyのmsgpackデコード失敗 | BadRequest | - | lobby/service/api.go: handleSearchRooms() | - |
| GameCacheからの取得失敗 | InternalServerError | - | lobby/room_cache.go: roomCacheQuery.do() | - |

※該当する部屋が無かった場合は、200 OKでroomsが空配列になります。このときResponseTypeはNoRoomFoundです。


## Search Rooms by Room IDs

POST /rooms/search/ids

### エラーレスポンス
| 概要 | HTTP Status (ResponseType) | gRPC Code | 発生箇所  | 備考 |
|------|----------------------------|-----------|-----------|------|
| レスポンスのmsgpackエンコード失敗 | InternalServerError | - | lobby/service/api.go: renderResponse() | - |
| ユーザ認証失敗 | Unauthorized | - | lobby/service/api.go: LobbyService.authUser() | - |
| リクエストbodyのmsgpackデコード失敗 | BadRequest | - | lobby/service/api.go: handleSearchByIds() | - |
| DBからの取得失敗 | InternalServerError | - | lobby/room.go: rs.SearchByIds() | - |

※該当する部屋が無かった場合は、200 OKでroomsが空配列になります。このときResponseTypeはNoRoomFoundです。


## Watch Room

POST /rooms/watch/id/{roomId}
POST /rooms/watch/number/{roomNumber}

### エラーレスポンス
| 概要 | HTTP Status (ResponseType) | gRPC Code | 発生箇所  | 備考 |
|------|----------------------------|-----------|-----------|------|
| レスポンスのmsgpackエンコード失敗 | InternalServerError | - | lobby/service/api.go: renderResponse() | - |
| ユーザ認証失敗 | Unauthorized | - | lobby/service/api.go: LobbyService.authUser() | - |
| リクエストbodyのmsgpackデコード失敗 | BadRequest | - | lobby/service/api.go: handleWatchRoom{,ByRoomNumber}() | - |
| RoomIDが空 | BadRequest | - | lobby/service/api.go: handleWatchRoom() | - |
| RoomNumberが空または0 | BadRequest | - | lobby/service/api.go: handleWatchRoomByNumber() | - |
| appIdのAppが無い | InternalServerError | - | lobby/room.go: RoomService.WatchBy{Id,Number}() | ユーザ認証失敗しているはずなので起こらない |
| 観戦可能なRoomが見つからない | **200 OK** (NoRoomFound) | - | lobby/room.go: RoomService.WatchBy{Id,Number}() | - |
| publicPropsのデコード失敗 | InternalServerError | - | obby/room.go: RoomService.WatchBy{Id,Number}() | - |
| プロパティクエリ条件に合致しない | **200 OK** (NoRoomFound) | - | lobby/room.go: RoomService.WatchBy{Id,Number}() | - |
| gameサーバ取得失敗 | InternalServerError | - | lobby/game_cache.go: GameCache.Get() | - |
| gRPC ClientをPoolから取得失敗 | InternalServerError | - | lobby/room.go: RoomService.watch() | - |
| gRPCタイムアウト | InternalServerError | DeadlineExceeded | lobby/room.go: RoomService.watch() | lobby側で設定したタイムアウト |
| appIdのAppが無い | InternalServerError | Internal | game/service/grpc.go: GameService.Watch() | ユーザ認証失敗しているはずなので起こらない |
| gRPCタイムアウト | InternalServerError | Deadlineexceeded | game/repository.go: Repository.joinRoom() | game側で設定したタイムアウト |
| Roomが既に消えた | **200 OK** (NoRoomFound) | NotFound | game/repository.go: Repository.joinRoom() | lobbyでのチェック後に消えたパターン |
| Watchableでない | **200 OK** (NoRoomFound) | FailedPrecondition | game/room.go: msgWatch() | lobbyでのチェック後に折られた |
| 既に入室済み | Conflict | AlreadyExists | game/room.go: msgWatch() | Playerとして既存も含む |
| Player PropsのUnmarshal失敗 | BadRequest | InvalidArgument | game/client.go: newClient() | - |

