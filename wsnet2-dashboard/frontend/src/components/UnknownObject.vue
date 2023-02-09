<script lang="ts">
import { defineComponent } from "vue";
import { render } from "../utils/render";

export default defineComponent({
  props: {
    data: null,
  },
  setup: (props) => {
    return () => render(props.data);
  },
});
</script>
