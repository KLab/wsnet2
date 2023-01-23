using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Net;

namespace WSNet2
{
    /// <summary>
    /// ネットワークアクセスの情報通知クラス
    /// </summary>
    /// WSNet2内部で通信が発生したことを通知する仕組みを提供する。
    /// NetworkInformer.SetCallback() で通知先を設定することで有効化。
    /// Debugビルドでのみ有効。
    public class NetworkInformer
    {
        /// <summary>
        ///   送受信データ情報
        /// </summary>
        [Serializable]
        public abstract class Info
        {
            /// <summary>送信データ本体のサイズ</summary>
            /// <remarks>各種ヘッダ情報は含まない（HTTP、WebSocket、TCP、IPなど）</remarks>
            public int BodySize;
        }

        /// <summary>
        ///   Lobbyへの送信データ情報（HTTP）
        /// </summary>
        [Serializable]
        public class LobbySendInfo : Info
        {
            /// <summary>リクエストURL</summary>
            public string URL;
        }

        /// <summary>
        ///   Lobbyからの受信データ情報（HTTP）
        /// </summary>
        [Serializable]
        public class LobbyReceiveInfo : Info
        {
            /// <summary>リクエストURL</summary>
            public string URL;

            /// <summary>レスポンスのStatusCode</summary>
            public HttpStatusCode StatusCode;
        }

        /// <summary>
        ///   Roomへの接続リクエスト送信（HTTP）
        /// </summary>
        [Serializable]
        public class RoomConnectRequestInfo : Info
        {
            /// <summary>リクエストURL</summary>
            public string URL;

            /// <summary>部屋ID</summary>
            public string RoomID;
        }

        /// <summary>
        ///   Roomへの送信データ情報（WebSocket）
        /// </summary>
        [Serializable]
        public abstract class RoomSendInfo : Info
        {
            /// <summary>部屋ID</summary>
            public string RoomID;

            /// <summary>メッセージ種別</summary>
            public MsgType MsgType;

            /// <summary>シーケンス番号</summary>
            /// <remarks>Pingでは0</remarks>
            public int SequenceNum;
        }

        /// <summary>
        ///   Ping送信情報
        /// </summary>
        [Serializable]
        public class RoomSendPingInfo : RoomSendInfo
        {
            /// <summary>タイムスタンプ（ミリ秒）</summary>
            public ulong TimestampMilliSec;
        }

        /// <summary>
        ///   退室要求送信情報
        /// </summary>
        [Serializable]
        public class RoomSendLeaveInfo : RoomSendInfo
        {
        }

        /// <summary>
        ///   Roomプロパティ変更送信情報
        /// </summary>
        [Serializable]
        public class RoomSendRoomPropInfo : RoomSendInfo
        {
            /// <summary>検索可能</summary>
            public bool Visible;

            /// <summary>入室可能</summary>
            public bool Joinable;

            /// <summary>観戦可能</summary>
            public bool Watchable;

            /// <summary>検索グループ</summary>
            public uint SearchGroup;

            /// <summary>最大人数</summary>
            public ushort MaxPlayers;

            /// <summary>通信タイムアウト時間(秒)</summary>
            public ushort ClientDeadline;

            /// <summary>公開プロパティ</summary>
            public byte[] PublicProps;

            /// <summary>非公開プロパティ</summary>
            public byte[] PrivateProps;
        }

        /// <summary>
        ///   プレイヤープロパティ変更送信情報
        /// </summary>
        [Serializable]
        public class RoomSendPlayerPropInfo : RoomSendInfo
        {
            /// <summary>プロパティ</summary>
            public byte[] Props;
        }

        /// <summary>
        ///   Master移譲送信情報
        /// </summary>
        [Serializable]
        public class RoomSendSwitchMasterInfo : RoomSendInfo
        {
            /// <summary>新Master</summary>
            public string NewMaster;
        }

        /// <summary>
        ///   RPC送信情報
        /// </summary>
        /// MsgType: Msg.Target, Msg.ToMaster, Msg.Broadcast
        [Serializable]
        public class RoomSendRPCInfo : RoomSendInfo
        {
            /// <summary>Roomに登録されたRPC ID</summary>
            public int RpcID;

            /// <summary>メソッド名</summary>
            public string MethodName;

            /// <summary>送信対象</summary>
            /// <remarks>
            ///   Msg.ToMasterではnull、Msg.Broadcastでは空配列
            /// </remarks>
            public string[] Targets;

            /// <summary>パラメータ</summary>
            public byte[] Param;
        }

        /// <summary>
        ///   Kick送信情報
        /// </summary>
        [Serializable]
        public class RoomSendKickInfo : RoomSendInfo
        {
            /// <summary>ターゲット</summary>
            public string Target;
        }

        /// <summary>
        ///   Roomからの受信データ情報（WebSocket）
        /// </summary>
        [Serializable]
        public abstract class RoomReceiveInfo : Info
        {
            /// <summary>部屋ID</summary>
            public string RoomID;

            /// <summary>イベント種別</summary>
            public EvType EvType;

            /// <summary>シーケンス番号</summary>
            /// <remarks>Pong, PeerReadyでは0</remarks>
            public int SequenceNum;
        }

        /// <summary>
        ///   PeerReady受信情報
        /// </summary>
        [Serializable]
        public class RoomReceivePeerReadyInfo : RoomReceiveInfo
        {
            /// <summary>Gameサーバが最後に受取ったMsgの通し番号</summary>
            public int LastMsgSeqNum;
        }

        /// <summary>
        ///   Pong受信情報
        /// </summary>
        [Serializable]
        public class RoomReceivePongInfo : RoomReceiveInfo
        {
            /// <summary>PingTimestamp</summary>
            public ulong PingTimestampMilliSec;

            /// <summary>RTT</summary>
            public ulong RTT;

            /// <summary>観戦者数</summary>
            public uint WatcherCount;

            /// <summary>各Playerの最終Msg時刻</summary>
            public IReadOnlyDictionary<string, ulong> LastMsgTimestamps;
        }

        /// <summary>
        ///   入室/再入室通知受信情報
        /// </summary>
        /// <remarks>EvType: Joined or Rejoined</remarks>
        [Serializable]
        public class RoomReceiveJoinedInfo : RoomReceiveInfo
        {
            /// <summary>入室/再入室したPlayerのID</summary>
            public string PlayerID;

            /// <summary>Playerのプロパティ</summary>
            public byte[] Props;
        }

        /// <summary>
        ///   退室通知受信情報ev
        /// </summary>
        [Serializable]
        public class RoomReceiveLeftInfo : RoomReceiveInfo
        {
            /// <summary>退室したPlayerのID</summary>
            public string PlayerID;

            /// <summary>現在のMasterID</summary>
            public string MasterID;
        }

        /// <summary>
        ///   部屋プロパティ変更通知受信情報
        /// </summary>
        [Serializable]
        public class RoomReceiveRoomPropInfo : RoomReceiveInfo
        {
            /// <summary>検索可能</summary>
            public bool Visible;

            /// <summary>入室可能</summary>
            public bool Joinable;

            /// <summary>観戦可能</summary>
            public bool Watchable;

            /// <summary>検索グループ</summary>
            public uint SearchGroup;

            /// <summary>最大プレイヤー数</summary>
            public ushort MaxPlayers;

            /// <summary>通信タイムアウト時間(秒)</summary>
            public ushort ClientDeadline;

            /// <summary>公開プロパティ</summary>
            public byte[] PublicProps;

            /// <summary>非公開プロパティ</summary>
            public byte[] PrivateProps;
        }

        /// <summary>
        ///   プレイヤープロパティ変更通知受信情報
        /// </summary>
        [Serializable]
        public class RoomReceivePlayerPropInfo : RoomReceiveInfo
        {
            /// <summary>プレイヤーID</summary>
            public string PlayerID;

            /// <summary>プロパティ</summary>
            public byte[] Props;
        }

        /// <summary>
        ///   Master交代通知受信情報
        /// </summary>
        [Serializable]
        public class RoomReceiveMasterSwitchedInfo : RoomReceiveInfo
        {
            /// <summary>新マスターID</summary>
            public string NewMasterID;
        }

        /// <summary>
        ///   RPC受信情報
        /// </summary>
        [Serializable]
        public class RoomReceiveRPCInfo : RoomReceiveInfo
        {
            /// <summary>送信者</summary>
            public string Sender;

            /// <summary>Roomに登録されたRPC ID</summary>
            public int RpcID;

            /// <summary>メソッド名</summary>
            public string MethodName;

            /// <summary>パラメータ</summary>
            public byte[] Param;
        }

        /// <summary>
        ///   レスポンス受信情報
        /// </summary>
        [Serializable]
        public class RoomReceiveResponseInfo : RoomReceiveInfo
        {
            /// <summary>元となるMsgのシーケンス番号</summary>
            public int MsgSeqNum;

            /// <summary>TargetNotFoundのとき不在だったTarget</summary>
            public string[] Targets;
        }

#if DEBUG
        static Action<Info> callback = null;
        static object callbackLock = new object();

        /// <summary>
        ///   readerから1要素分の配列を切り出す
        /// </summary>
        public static byte[] CutOutOne(SerialReader reader)
        {
            var rest = reader.GetRest();
            _ = reader.Read();
            var count = rest.Count - reader.GetRest().Count;
            var data = new byte[count];
            Buffer.BlockCopy(rest.Array, rest.Offset, data, 0, count);
            return data;
        }
#endif

        /// <summary>
        ///   通知を受け取るCallbackを設定
        /// </summary>
        /// <param name="callback">通知を受け取るCallback</param>
        /// <remarks>
        ///   callbackをnullにすることで無効化できる
        /// </remarks>
        [Conditional("DEBUG")]
        public static void SetCallback(Action<Info> callback)
        {
#if DEBUG
            lock (callbackLock)
            {
                NetworkInformer.callback = callback;
            }
#endif
        }

        /// <summary>
        ///   Lobbyへの送信発生（内部用）
        /// </summary>
        [Conditional("DEBUG")]
        public static void OnLobbySend(string url, byte[] body)
        {
#if DEBUG
            lock (callbackLock)
            {
                callback?.Invoke(
                    new LobbySendInfo()
                    {
                        BodySize = body.Length,
                        URL = url,
                    });
            }
#endif
        }

        /// <summary>
        ///   Lobbyからの受信発生（内部用）
        /// </summary>
        [Conditional("DEBUG")]
        public static void OnLobbyReceive(string url, HttpStatusCode code, byte[] body)
        {
#if DEBUG
            lock (callbackLock)
            {
                callback?.Invoke(
                    new LobbyReceiveInfo()
                    {
                        BodySize = body.Length,
                        URL = url,
                        StatusCode = code,
                    });
            }
#endif
        }

        /// <summary>
        ///   Roomへの接続リクエスト発生（内部用）
        /// </summary>
        [Conditional("DEBUG")]
        public static void OnRoomConnectRequest(Room room, string url)
        {
#if DEBUG
            lock (callbackLock)
            {
                callback?.Invoke(
                    new RoomConnectRequestInfo()
                    {
                        URL = url,
                        RoomID = room.Id,
                    });
            }
#endif
        }

        /// <summary>
        ///   Roomへの送信発生（内部用）
        /// </summary>
        [Conditional("DEBUG")]
        public static void OnRoomSend(Room room, ArraySegment<byte> payload)
        {
#if DEBUG
            lock (callbackLock)
            {
                if (callback == null)
                {
                    return;
                }

                callback(MsgPool.ParsePayload(room, payload));
            }
#endif
        }

        /// <summary>
        ///   Roomからの受信発生（内部用）
        /// </summary>
        [Conditional("DEBUG")]
        public static void OnRoomReceive(Room room, int bodySize, Event ev)
        {
#if DEBUG
            lock (callbackLock)
            {
                if (callback == null)
                {
                    return;
                }

                RoomReceiveInfo info;

                switch (ev)
                {
                    case EvPeerReady evPeerReady:
                        info = new RoomReceivePeerReadyInfo()
                        {
                            BodySize = bodySize,
                            RoomID = room.Id,
                            EvType = ev.Type,
                            SequenceNum = 0,
                            LastMsgSeqNum = evPeerReady.LastMsgSeqNum,
                        };
                        break;
                    case EvPong evPong:
                        var lastMsgTimestamps = new Dictionary<string, ulong>();
                        evPong.GetLastMsgTimestamps(lastMsgTimestamps);
                        info = new RoomReceivePongInfo()
                        {
                            BodySize = bodySize,
                            RoomID = room.Id,
                            EvType = ev.Type,
                            SequenceNum = 0,
                            PingTimestampMilliSec = evPong.PingTimestamp,
                            RTT = evPong.RTT,
                            WatcherCount = evPong.WatcherCount,
                            LastMsgTimestamps = lastMsgTimestamps,
                        };
                        break;
                    case EvJoined evJoined:
                        info = new RoomReceiveJoinedInfo()
                        {
                            BodySize = bodySize,
                            RoomID = room.Id,
                            EvType = ev.Type,
                            PlayerID = evJoined.ClientID,
                            Props = CutOutOne(evJoined.GetUnread()),
                        };
                        break;
                    case EvRejoined evRejoined:
                        info = new RoomReceiveJoinedInfo()
                        {
                            BodySize = bodySize,
                            RoomID = room.Id,
                            EvType = ev.Type,
                            PlayerID = evRejoined.ClientID,
                            Props = CutOutOne(evRejoined.GetUnread()),
                        };
                        break;
                    case EvLeft evLeft:
                        info = new RoomReceiveLeftInfo()
                        {
                            BodySize = bodySize,
                            RoomID = room.Id,
                            EvType = ev.Type,
                            PlayerID = evLeft.ClientID,
                            MasterID = evLeft.MasterID,
                        };
                        break;
                    case EvRoomProp evRoomProp:
                        var reader = ev.GetUnread();
                        info = new RoomReceiveRoomPropInfo()
                        {
                            BodySize = bodySize,
                            RoomID = room.Id,
                            EvType = ev.Type,
                            Visible = evRoomProp.Visible,
                            Joinable = evRoomProp.Joinable,
                            Watchable = evRoomProp.Watchable,
                            SearchGroup = evRoomProp.SearchGroup,
                            MaxPlayers = evRoomProp.MaxPlayers,
                            ClientDeadline = evRoomProp.ClientDeadline,
                            PublicProps = CutOutOne(reader),
                            PrivateProps = CutOutOne(reader),
                        };
                        break;
                    case EvClientProp evClientProp:
                        info = new RoomReceivePlayerPropInfo()
                        {
                            BodySize = bodySize,
                            RoomID = room.Id,
                            EvType = ev.Type,
                            PlayerID = evClientProp.ClientID,
                            Props = CutOutOne(evClientProp.GetUnread()),
                        };
                        break;
                    case EvMasterSwitched evMasterSwitched:
                        info = new RoomReceiveMasterSwitchedInfo()
                        {
                            BodySize = bodySize,
                            RoomID = room.Id,
                            EvType = ev.Type,
                            NewMasterID = evMasterSwitched.NewMasterId,
                        };
                        break;
                    case EvRPC evRpc:
                        info = new RoomReceiveRPCInfo()
                        {
                            BodySize = bodySize,
                            RoomID = room.Id,
                            EvType = ev.Type,
                            Sender = evRpc.SenderID,
                            RpcID = evRpc.RpcID,
                            MethodName = room.MethodName(evRpc.RpcID),
                            Param = CutOutOne(evRpc.GetUnread()),
                        };
                        break;
                    case EvResponse evResponse:
                        info = new RoomReceiveResponseInfo()
                        {
                            BodySize = bodySize,
                            RoomID = room.Id,
                            EvType = ev.Type,
                            MsgSeqNum = evResponse.MsgSeqNum,
                            Targets = evResponse.Targets,
                        };
                        break;
                    default:
                        throw new Exception($"unknown event: {ev}");
                }

                callback(info);
            }
#endif
        }
    }
}
