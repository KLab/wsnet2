using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Threading.Tasks;

namespace WSNet2
{
    /// <summary>
    ///   System.Net.Http.HttpClientを使ったPOSTの実装
    /// </summary>
    static class DefaultHttpClient
    {
        static HttpClient client;

        public static void Post(string url, IReadOnlyDictionary<string, string> headers, byte[] content, TaskCompletionSource<(int, byte[])> tcs)
        {
            Task.Run(async () =>
            {
                try
                {
                    var request = new HttpRequestMessage(HttpMethod.Post, url);
                    request.Content = new ByteArrayContent(content);
                    foreach (var kv in headers)
                    {
                        request.Headers.Add(kv.Key, kv.Value);
                    }

                    client ??= new HttpClient();
                    var res = await client.SendAsync(request);
                    var body = await res.Content.ReadAsByteArrayAsync();

                    tcs.TrySetResult(((int)res.StatusCode, body));
                }
                catch (Exception e)
                {
                    tcs.TrySetException(e);
                }
            });
        }
    }
}
