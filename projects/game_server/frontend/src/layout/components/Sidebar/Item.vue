<script>
import { h } from 'vue'

export default {
  name: 'MenuItem',
  props: {
    icon: {
      type: String,
      default: ''
    },
    title: {
      type: String,
      default: ''
    }
  },
  render() {
    const vnodes = []

    if (this.icon) {
      vnodes.push(h('svg-icon', { 'icon-class': this.icon }))
    }

    if (this.title) {
      vnodes.push(h('span', { slot: 'title' }, this.title))
    }

    return vnodes
  }
}
</script>
