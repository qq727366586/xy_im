package timer

import (
	"sync"
	"time"
)

const (
	timerFormat      = "2006-01-02 15:04:05"
	infiniteDuration = time.Duration(1<<63 - 1)
)

// 最小堆定时器
type Timer struct {
	lock   sync.Mutex
	signal *time.Timer
	free   *TimerData
	timers []*TimerData // 最小堆存储结构
	num    int
}

type TimerData struct {
	Key    string
	index  int
	expire time.Time
	next   *TimerData
	fn     func()
}

func (td *TimerData) Delay() time.Duration {
	return time.Until(td.expire)
}

func (td *TimerData) ExpireString() string {
	return td.expire.Format(timerFormat)
}
func NewTimer(num int) (t *Timer) {
	t = new(Timer)
	t.init(num)
	return t
}

func (t *Timer) Init(num int) {
	t.init(num)
}

// 初始化
func (t *Timer) init(num int) {
	t.signal = time.NewTimer(infiniteDuration)
	t.timers = make([]*TimerData, 0, num)
	t.num = num
	t.grow()
	go t.start()
}

// 开启监听
func (t *Timer) start() {
	for {
		t.expire()
		<-t.signal.C
	}
}

func (t *Timer) expire() {
	var (
		fn func()
		td *TimerData
		d  time.Duration
	)
	t.lock.Lock()
	for {
		if len(t.timers) == 0 {
			d = infiniteDuration
			break
		}
		td = t.timers[0]
		if d = td.Delay(); d > 0 {
			break
		}
		fn = td.fn
		// 让调用者选择是否归还
		t.del(td)
		t.lock.Unlock()
		if fn != nil {
			fn()
		}
		t.lock.Lock()
	}
	t.signal.Reset(d)
	t.lock.Unlock()
}

// 预生成空闲数据
func (t *Timer) grow() {
	var (
		i   int
		td  *TimerData
		tds = make([]TimerData, t.num)
	)
	t.free = &(tds[0])
	td = t.free
	for i = 1; i < t.num; i++ {
		td.next = &(tds[i])
		td = td.next
	}
	td.next = nil
}

func (t *Timer) get() (td *TimerData) {
	if td = t.free; td == nil {
		t.grow()
		td = t.free
	}
	t.free = td.next
	return
}

func (t *Timer) Add(expire time.Duration, fn func()) (td *TimerData) {
	t.lock.Lock()
	td = t.get()
	td.expire = time.Now().Add(expire)
	td.fn = fn
	t.add(td)
	t.lock.Unlock()
	return
}

func (t *Timer) add(td *TimerData) {
	td.index = len(t.timers)
	t.timers = append(t.timers, td)
	t.up(td.index)
	// 如果第一个,需要立即唤醒
	if td.index == 0 {
		d := td.Delay()
		t.signal.Reset(d)
	}
}

// 删除
func (t *Timer) Del(td *TimerData) {
	t.lock.Lock()
	t.del(td)
	t.put(td)
	t.lock.Unlock()
}

func (t *Timer) del(td *TimerData) {
	i := td.index
	last := len(t.timers) - 1
	// 可能已经移除了, 通常是已经过期了
	if i < 0 || i > last || t.timers[i] != td {
		return
	}
	if i != last {
		t.swap(i, last)
		t.down(i, last)
		t.up(i)
	}
	// 安全起见,设置最后一个 index -1
	t.timers[last].index = -1
	t.timers = t.timers[:last]
}

func (t *Timer) put(td *TimerData) {
	td.fn = nil
	td.next = t.free
	t.free = td
}

// 重新设置时间
func (t *Timer) Set(td *TimerData, expire time.Duration) {
	t.lock.Lock()
	t.del(td)
	td.expire = time.Now().Add(expire)
	t.add(td)
	t.lock.Unlock()
}

func (t *Timer) up(j int) {
	for {
		i := (j - 1) / 2
		if i >= j || !t.less(j, i) {
			break
		}
		t.swap(i, j)
		j = i
	}
}

func (t *Timer) down(i, n int) {
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 {
			break
		}
		j := j1
		if j2 := j1 + 1; j2 < n && !t.less(j1, j2) {
			j = j2
		}
		if !t.less(j, i) {
			break
		}
		t.swap(i, j)
		i = j
	}
}

func (t *Timer) less(i, j int) bool {
	return t.timers[i].expire.Before(t.timers[j].expire)
}

func (t *Timer) swap(i, j int) {
	t.timers[i], t.timers[j] = t.timers[j], t.timers[i]
	t.timers[i].index = i
	t.timers[j].index = j
}
