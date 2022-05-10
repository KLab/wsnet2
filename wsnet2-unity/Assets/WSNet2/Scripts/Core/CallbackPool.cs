using System;
using System.Collections.Concurrent;

namespace WSNet2
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
        Func<bool> isRunning;

        public CallbackPool()
        {
            isRunning = () => true;
        }

        public CallbackPool(Func<bool> isRunning)
        {
            this.isRunning = isRunning;
        }

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
            while (isRunning())
            {
                Action callback;

                if (!queue.TryDequeue(out callback))
                {
                    return;
                }

                callback();
            }
        }
    }
}
