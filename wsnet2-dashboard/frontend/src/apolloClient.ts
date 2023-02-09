// graphql
import {
  ApolloClient,
  createHttpLink,
  InMemoryCache,
} from "@apollo/client/core";

// graphql
export default new ApolloClient({
  cache: new InMemoryCache(),
});
