<template>
  <a-modal
    v-model:open="open"
    class="budget-edit-modal"
    :title="t('result.budget.editModalTitle')"
    :width="440"
    :confirm-loading="saving"
    :ok-text="t('common.save')"
    :cancel-text="t('common.cancel')"
    :ok-button-props="{ disabled: value === null }"
    @ok="emit('confirm')"
    @cancel="emit('cancel')"
  >
    <div
      v-if="target"
      class="budget-edit-body"
      :class="[
        `budget-edit-body--${pandaMood}`,
        { 'budget-edit-body--celebrating': celebrating },
      ]"
    >
      <div class="budget-edit-panda" :class="`budget-edit-panda--${pandaMood}`" aria-hidden="true">
        <span class="budget-edit-panda__ears">
          <span class="budget-edit-panda__ear budget-edit-panda__ear--left"></span>
          <span class="budget-edit-panda__ear budget-edit-panda__ear--right"></span>
        </span>
        <span class="budget-edit-panda__face">
          <span class="budget-edit-panda__eye budget-edit-panda__eye--left"><i></i></span>
          <span class="budget-edit-panda__eye budget-edit-panda__eye--right"><i></i></span>
          <span class="budget-edit-panda__nose"></span>
          <span
            class="budget-edit-panda__mouth"
            :class="{ 'budget-edit-panda__mouth--sad': pandaMood === 'sad' }"
          ></span>
          <span v-if="pandaMood === 'happy'" class="budget-edit-panda__blush budget-edit-panda__blush--left"></span>
          <span v-if="pandaMood === 'happy'" class="budget-edit-panda__blush budget-edit-panda__blush--right"></span>
          <span v-if="pandaMood === 'sad'" class="budget-edit-panda__tear budget-edit-panda__tear--left"></span>
          <span v-if="pandaMood === 'sad'" class="budget-edit-panda__tear budget-edit-panda__tear--right"></span>
          <span class="budget-edit-panda__brow budget-edit-panda__brow--left"></span>
          <span class="budget-edit-panda__brow budget-edit-panda__brow--right"></span>
        </span>
        <span class="budget-edit-panda__body">
          <span class="budget-edit-panda__arm budget-edit-panda__arm--left"></span>
          <span class="budget-edit-panda__belly">¥</span>
          <span class="budget-edit-panda__arm budget-edit-panda__arm--right"></span>
        </span>
        <span class="budget-edit-panda__bamboo"></span>
        <span v-if="celebrating" class="budget-edit-confetti">
          <span v-for="n in 10" :key="n" class="budget-edit-confetti__leaf" :class="`budget-edit-confetti__leaf--${n}`"></span>
          <span v-for="n in 6" :key="`c-${n}`" class="budget-edit-confetti__coin" :class="`budget-edit-confetti__coin--${n}`">¥</span>
        </span>
      </div>
      <p class="budget-edit-panda__mood">{{ pandaMoodText }}</p>
      <p class="budget-edit-name">{{ target.name }}</p>
      <p class="budget-edit-meta">
        <span class="budget-edit-type">{{ typeLabel }}</span>
        <span v-if="target.dayNumber" class="budget-edit-day">{{ t('common.dayNumber', { day: target.dayNumber }) }}</span>
        <span class="budget-edit-current">{{ t('result.budget.editModalCurrent', { amount: formatAmount(target.amount) }) }}</span>
      </p>
      <a-form-item :label="t('result.budget.editModalLabel')" class="budget-edit-field">
        <a-input-number
          :value="value"
          class="budget-edit-input"
          :min="0"
          :precision="2"
          :step="10"
          :placeholder="t('result.budget.editModalPlaceholder')"
          addon-before="¥"
          style="width: 100%"
          @update:value="emit('update:value', $event)"
          @press-enter="emit('confirm')"
        />
      </a-form-item>
      <div v-if="value !== null && delta !== 0" class="budget-edit-delta" :class="delta > 0 ? 'budget-edit-delta--up' : 'budget-edit-delta--down'">
        {{ t('result.budget.editModalDelta', { amount: formatAmount(Math.abs(delta)) }) }}
        {{ delta > 0 ? t('result.budget.editModalDeltaUp') : t('result.budget.editModalDeltaDown') }}
      </div>
      <p v-else class="budget-edit-hint">{{ t('result.budget.editModalHint') }}</p>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

type BudgetTarget = {
  name: string
  amount: number
  type: string
  dayNumber: number | null
}

const props = withDefaults(defineProps<{
  open: boolean
  target: BudgetTarget | null
  value: number | null
  saving: boolean
  celebrating: boolean
  typeLabel: string
  formatAmount: (value: number) => string
}>(), {
  target: null,
  value: null,
  saving: false,
  celebrating: false,
  typeLabel: '',
})

const emit = defineEmits<{
  'update:open': [value: boolean]
  'update:value': [value: number | null]
  confirm: []
  cancel: []
}>()

const { t } = useI18n()
const delta = computed(() => {
  if (!props.target || props.value === null) return 0
  return Math.round((props.value - props.target.amount) * 100) / 100
})
const pandaMood = computed<'happy' | 'sad' | 'idle'>(() => {
  if (delta.value > 0) return 'sad'
  if (delta.value < 0) return 'happy'
  return 'idle'
})
const pandaMoodText = computed(() => {
  if (pandaMood.value === 'happy') return t('result.budget.editModalPandaHappy')
  if (pandaMood.value === 'sad') return t('result.budget.editModalPandaSad')
  return t('result.budget.editModalPandaIdle')
})

const open = computed({
  get: () => props.open,
  set: (value: boolean) => emit('update:open', value),
})
</script>

<style scoped>
.budget-edit-modal :deep(.ant-modal-body) { padding-top: 8px; }
</style>
.budget-edit-body { position: relative; display: flex; flex-direction: column; gap: 10px; overflow: hidden; }
.budget-edit-panda { position: relative; width: 92px; height: 104px; margin: 2px auto 0; flex: none; transition: transform .3s ease; }
.budget-edit-panda__ears { position: absolute; left: 12px; right: 12px; top: 0; height: 22px; z-index: 1; }
.budget-edit-panda__ear { position: absolute; top: 0; width: 22px; height: 22px; border-radius: 50%; background: #2e3a3d; transition: transform .3s ease; }
.budget-edit-panda__ear--left { left: 0; transform: rotate(-12deg); }
.budget-edit-panda__ear--right { right: 0; transform: rotate(12deg); }
.budget-edit-panda--happy .budget-edit-panda__ear--left { transform: rotate(-24deg); }
.budget-edit-panda--happy .budget-edit-panda__ear--right { transform: rotate(24deg); }
.budget-edit-panda--happy { animation: budget-panda-bounce .9s ease-in-out infinite; }
.budget-edit-panda__face { position: absolute; left: 8px; right: 8px; top: 8px; height: 62px; border-radius: 46% 46% 48% 48%; background: #fffdf8; border: 1px solid rgba(46,58,61,.12); box-shadow: inset 0 -6px 12px rgba(46,58,61,.06); z-index: 2; }
.budget-edit-panda__eye { position: absolute; top: 14px; width: 17px; height: 21px; border-radius: 50%; background: #2e3a3d; animation: budget-panda-blink 4.2s infinite; }
.budget-edit-panda__eye--left { left: 12px; transform: rotate(-14deg); }
.budget-edit-panda__eye--right { right: 12px; transform: rotate(14deg); }
.budget-edit-panda__eye i { position: absolute; left: 5px; top: 5px; width: 5px; height: 5px; border-radius: 50%; background: #fff; transition: transform .25s ease; }
.budget-edit-body--sad .budget-edit-panda__eye i { transform: translateY(-3px); }
.budget-edit-body--happy .budget-edit-panda__eye i { transform: translateY(3px); }
.budget-edit-body--sad .budget-edit-panda__eye { height: 18px; }
.budget-edit-body--happy .budget-edit-panda__eye { height: 23px; }
.budget-edit-panda__nose { position: absolute; left: 50%; top: 34px; width: 9px; height: 7px; border-radius: 50%; background: #2e3a3d; transform: translateX(-50%); }
.budget-edit-panda__mouth { position: absolute; left: 50%; top: 41px; width: 15px; height: 8px; border: 2px solid #2e3a3d; border-top: 0; border-left-color: transparent; border-right-color: transparent; border-radius: 0 0 15px 15px; transform: translateX(-50%); }
.budget-edit-panda--happy .budget-edit-panda__mouth { width: 20px; height: 11px; }
.budget-edit-panda__mouth--sad { top: 43px; border-radius: 15px 15px 0 0; border: 2px solid #2e3a3d; border-bottom: 0; border-left-color: transparent; border-right-color: transparent; }
.budget-edit-panda__brow { position: absolute; top: 6px; width: 14px; height: 3px; border-radius: 999px; background: #2e3a3d; opacity: 0; }
.budget-edit-panda__brow--left { left: 10px; }
.budget-edit-panda__brow--right { right: 10px; }
.budget-edit-panda--happy .budget-edit-panda__brow, .budget-edit-panda--sad .budget-edit-panda__brow { opacity: 1; }
.budget-edit-panda--happy .budget-edit-panda__brow--left { transform: rotate(-10deg); }
.budget-edit-panda--happy .budget-edit-panda__brow--right { transform: rotate(10deg); }
.budget-edit-panda--sad .budget-edit-panda__brow--left { transform: rotate(12deg); }
.budget-edit-panda--sad .budget-edit-panda__brow--right { transform: rotate(-12deg); }
.budget-edit-panda__blush { position: absolute; top: 34px; width: 11px; height: 7px; border-radius: 50%; background: rgba(184,69,62,.3); }
.budget-edit-panda__blush--left { left: 6px; }
.budget-edit-panda__blush--right { right: 6px; }
.budget-edit-panda__tear { position: absolute; top: 32px; width: 5px; height: 8px; border-radius: 50%; background: rgba(93,164,177,.85); animation: budget-panda-tear 1.4s ease-in infinite; }
.budget-edit-panda__tear--left { left: 14px; }
.budget-edit-panda__tear--right { right: 14px; animation-delay: .35s; }
.budget-edit-panda__body { position: absolute; left: 18px; right: 18px; bottom: 12px; height: 36px; z-index: 1; }
.budget-edit-panda__belly { position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; border-radius: 48% 48% 46% 46%; background: #fffdf8; border: 1px solid rgba(46,58,61,.12); color: rgba(184,69,62,.55); font-size: 18px; font-weight: 800; }
.budget-edit-panda__arm { position: absolute; top: 2px; width: 15px; height: 26px; border-radius: 999px; background: #2e3a3d; transform-origin: top center; animation: budget-panda-wave 2.6s ease-in-out infinite; }
.budget-edit-panda__arm--left { left: -7px; }
.budget-edit-panda__arm--right { right: -7px; animation-delay: .5s; }
.budget-edit-panda--happy .budget-edit-panda__arm { animation-duration: .9s; }
.budget-edit-panda__bamboo { position: absolute; right: 8px; bottom: 14px; width: 8px; height: 46px; border-radius: 999px; background: linear-gradient(180deg,#6a7a6a,#4c5b4c); transform: rotate(14deg); z-index: 0; }
.budget-edit-panda__bamboo::before, .budget-edit-panda__bamboo::after { content: ''; position: absolute; left: -3px; width: 14px; height: 8px; border-radius: 999px; background: #7d927d; }
.budget-edit-panda__bamboo::before { top: 8px; transform: rotate(-18deg); }
.budget-edit-panda__bamboo::after { top: 22px; transform: rotate(18deg); }
.budget-edit-body--celebrating .budget-edit-panda { animation: budget-panda-celebrate .55s ease-in-out 2; }
.budget-edit-confetti { position: absolute; left: 50%; top: -6px; width: 0; height: 0; pointer-events: none; z-index: 5; }
.budget-edit-confetti__leaf, .budget-edit-confetti__coin { position: absolute; left: 0; top: 0; opacity: 0; animation: budget-confetti-fall 1.1s ease-out forwards; }
.budget-edit-confetti__leaf { width: 9px; height: 14px; border-radius: 60% 10% 60% 10%; background: #7d927d; }
.budget-edit-confetti__coin { display: flex; align-items: center; justify-content: center; width: 16px; height: 16px; border-radius: 50%; background: #c9a96e; border: 1px solid #a8843c; color: #fffdf8; font-size: 10px; font-weight: 800; }
.budget-edit-confetti__leaf--1 { --tx: -64px; --r: -70deg; }
.budget-edit-confetti__leaf--2 { --tx: -42px; --r: 50deg; animation-delay: .05s; }
.budget-edit-confetti__leaf--3 { --tx: -20px; --r: -40deg; animation-delay: .1s; }
.budget-edit-confetti__leaf--4 { --tx: 12px; --r: 80deg; animation-delay: .02s; }
.budget-edit-confetti__leaf--5 { --tx: 36px; --r: -60deg; animation-delay: .08s; }
.budget-edit-confetti__leaf--6 { --tx: 58px; --r: 40deg; animation-delay: .03s; }
.budget-edit-confetti__leaf--7 { --tx: -56px; --r: -80deg; animation-delay: .06s; }
.budget-edit-confetti__leaf--8 { --tx: 50px; --r: 60deg; animation-delay: .11s; }
.budget-edit-confetti__leaf--9 { --tx: -30px; --r: -50deg; animation-delay: .03s; }
.budget-edit-confetti__leaf--10 { --tx: 28px; --r: 75deg; animation-delay: .09s; }
.budget-edit-confetti__coin--1 { --tx: -40px; --r: 180deg; animation-delay: .04s; }
.budget-edit-confetti__coin--2 { --tx: -18px; --r: -180deg; animation-delay: .1s; }
.budget-edit-confetti__coin--3 { --tx: 2px; --r: 180deg; }
.budget-edit-confetti__coin--4 { --tx: 24px; --r: -180deg; animation-delay: .07s; }
.budget-edit-confetti__coin--5 { --tx: 44px; --r: 180deg; animation-delay: .02s; }
.budget-edit-confetti__coin--6 { --tx: -56px; --r: -180deg; animation-delay: .12s; }
.budget-edit-panda__mood { margin: 0; text-align: center; font-size: 12px; letter-spacing: .06em; color: #8a9a9e; }
.budget-edit-body--happy .budget-edit-panda__mood { color: #3e7d8a; }
.budget-edit-body--sad .budget-edit-panda__mood { color: #b8453e; }
.budget-edit-name { margin: 0; font-size: 16px; font-weight: 700; color: #2e3a3d; line-height: 1.5; }
.budget-edit-meta { margin: 0; display: flex; flex-wrap: wrap; align-items: center; gap: 8px; font-size: 13px; color: #8a9a9e; }
.budget-edit-type { padding: 2px 10px; border-radius: 999px; background: rgba(93,164,177,.12); border: 1px solid rgba(93,164,177,.35); color: #3e7d8a; font-size: 12px; font-weight: 600; }
.budget-edit-current { font-weight: 600; color: #5e6e72; }
.budget-edit-field { margin-bottom: 0; }
.budget-edit-input :deep(.ant-input-number-input) { font-size: 18px; font-weight: 700; color: #b8453e; }
.budget-edit-delta { margin: 0; padding: 8px 12px; border-radius: 10px; font-size: 13px; font-weight: 600; }
.budget-edit-delta--up { background: rgba(184,69,62,.08); border: 1px solid rgba(184,69,62,.3); color: #b8453e; }
.budget-edit-delta--down { background: rgba(93,164,177,.08); border: 1px solid rgba(93,164,177,.35); color: #3e7d8a; }
.budget-edit-hint { margin: 0; font-size: 12px; line-height: 1.6; color: #8a9a9e; }

@keyframes budget-panda-blink { 0%,93%,100% { transform: rotate(-14deg) scaleY(1); } 95% { transform: rotate(-14deg) scaleY(.12); } }
@keyframes budget-panda-tear { 0% { transform: translateY(0); opacity: 0; } 25% { opacity: 1; } 100% { transform: translateY(8px); opacity: 0; } }
@keyframes budget-panda-wave { 0%,100% { transform: rotate(8deg); } 50% { transform: rotate(-16deg); } }
@keyframes budget-panda-bounce { 0%,100% { transform: translateY(0); } 50% { transform: translateY(-5px); } }
@keyframes budget-panda-celebrate { 0%,100% { transform: rotate(0) translateY(0); } 25% { transform: rotate(-7deg) translateY(-5px); } 75% { transform: rotate(7deg) translateY(-5px); } }
@keyframes budget-confetti-fall { 0% { transform: translate(0,0) rotate(0); opacity: 1; } 100% { transform: translate(var(--tx,0),96px) rotate(var(--r,90deg)); opacity: 0; } }

.budget-edit-modal :deep(.ant-modal-body) { padding-top: 8px; }
