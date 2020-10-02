using System;
using System.Collections.Concurrent;

namespace WSNet2.Core
{
    /// <summary>
    ///   Callbackを溜めておいて後から実行できるようにするやつ.
    /// </summary>
    /// <remarks>
    ///   <para>
    ///     Unityではcallbackをメインスレッドで動かしたいので溜めておいて、
    ///     メインスレッドでProcess()を呼び出すようにする。
    ///   </para>
    ///   <para>
    ///     DotNetの場合にも適当なスレッドで定期的にProcess()を呼び出す必要がある。
    ///   </para>
    /// </remarks>
    public class CallbackPool
    {
        ConcurrentQueue<Action> queue = new ConcurrentQueue<Action>();

        /// <summary>
        ///   callbackをpoolに追加
        /// </summary>
        public void Add(Action callback)
        {
            queue.Enqueue(callback);
        }

        /// <summary>
        ///   Callbackを追加された順に実行する
        /// </summary>
        public void Process()
        {
            Action callback;
            while(queue.TryDequeue(out callback))
            {
                callback();
            }
        }
    }
}
