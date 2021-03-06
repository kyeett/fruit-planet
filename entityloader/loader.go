package entityloader

import (
	"fmt"
	"image"
	"log"
	"strconv"
	"strings"

	"github.com/kyeett/gomponents/direction"

	"github.com/fogleman/ease"

	"github.com/kyeett/fruit-planet/condition"
	"github.com/kyeett/gomponents/animation"

	"github.com/kyeett/gomponents/pathanimation"

	"github.com/hajimehoshi/ebiten/ebitenutil"

	"github.com/hajimehoshi/ebiten"
	"github.com/kyeett/ecs/entity"
	"github.com/kyeett/gomponents/components"
	tiled "github.com/lafriks/go-tiled"
	"github.com/peterhellberg/gfx"
)

var headless bool

var spritesheet *ebiten.Image

func init() {
	tmp, err := gfx.OpenPNG("assets/images/platformer.png")
	if err != nil {
		log.Fatal(err)
	}
	spritesheet, _ = ebiten.NewImageFromImage(tmp, ebiten.FilterDefault)
}

func Hitbox(em *entity.Manager, o *tiled.Object) {
	e := em.NewEntity("hitbox")
	em.Add(e, components.Pos{Vec: gfx.V(o.X, o.Y)})
	hitbox := components.NewHitbox(gfx.R(0, 0, o.Width, o.Height))
	if o.Properties.GetString("block_directions") != "" {
		hitbox.BlockedDirections = direction.FromString(o.Properties.GetString("block_directions"))
	}
	if o.Properties.GetBool("hazard") {
		em.Add(e, components.Hazard{})
	}
	em.Add(e, hitbox)
}

func Enemy(em *entity.Manager, o *tiled.Object) string {
	e := em.NewEntity("enemy")
	var hitbox gfx.Rect
	var subImage image.Rectangle
	switch o.Name {
	case "turnip":
		hitbox = gfx.R(0, 8, 16, 24)
		subImage = image.Rect(0, 0, 16, 24).Add(image.Pt(96, 32-8))
	case "beetroot":
		hitbox = gfx.R(3, 8, 13, 16)
		subImage = image.Rect(0, 0, 16, 16).Add(image.Pt(192-16, 24+32+8))
	}
	em.Add(e, components.Pos{Vec: gfx.V(o.X, o.Y-hitbox.Max.Y)})
	em.Add(e, components.Velocity{Vec: gfx.V(0, 0)})
	em.Add(e, components.NewHitbox(hitbox))
	if o.Properties.GetBool("hazard") {
		em.Add(e, components.Hazard{})
	}
	em.Add(e, components.Drawable{spritesheet.SubImage(subImage).(*ebiten.Image)})

	if o.Properties.GetString("on_path") != "" {
		em.Add(e, components.OnPath{
			Label:     o.Properties.GetString("on_path"),
			Speed:     1,
			Target:    1,
			Mode:      pathanimation.LinearLoop,
			Direction: 1,
		})
	}
	return e
}

func Area(em *entity.Manager, o *tiled.Object) {
	e := em.NewEntity("area")
	hitbox := gfx.R(o.X, o.Y, o.X+o.Width, o.Y+o.Height)
	em.Add(e, components.Area{
		Rect: hitbox,
		Name: o.Name,
	})
}

func Player(em *entity.Manager, o *tiled.Object) string {
	// Save initial position
	initial := em.NewEntity("initial")
	em.Add(initial, components.Pos{Vec: gfx.V(o.X, o.Y)})

	e := em.NewEntity("player")
	hitbox := gfx.R(4, 8, 18, 22)

	em.Add(e, components.Pos{Vec: gfx.V(o.X, o.Y)})
	em.Add(e, components.Velocity{Vec: gfx.V(0, 0)})
	em.Add(e, components.Joystick{})
	if !headless {
		tmp, err := gfx.OpenPNG("assets/images/platformer.png")
		if err != nil {
			log.Fatal(err)
		}
		pImage, _ := ebiten.NewImageFromImage(tmp, ebiten.FilterDefault)
		em.Add(e, components.Drawable{pImage.SubImage(image.Rect(5, 10, 27, 32)).(*ebiten.Image)})
	}
	em.Add(e, components.NewHitbox(hitbox))
	return e
}

func parseInAreaCondition(em *entity.Manager, prop *tiled.Property) condition.InArea {
	params := strings.Split(prop.Value, ",")

	// Find area matching name
	for _, e := range em.FilteredEntities(components.AreaType) {
		if em.Area(e).Name == params[1] {
			return condition.NewInArea(em, params[0], e)
		}
	}

	log.Fatal("no area exists with name", params[1])
	return condition.InArea{}
}

func Condition(em *entity.Manager, o *tiled.Object) {
	e := em.NewEntity("condition")
	cond := components.Trigger{
		Name: o.Name,
	}
	for _, p := range o.Properties {
		switch p.Name {
		case "key_pressed":
			cond.Conditions = append(cond.Conditions, condition.KeyPressed{Key: condition.KeyNameToKey(p.Value)})
		case "in_area":
			cond.Conditions = append(cond.Conditions, parseInAreaCondition(em, p))
		default:
			fmt.Println("Unknown condition property", o)
		}
	}
	em.Add(e, cond)
}

func Text(em *entity.Manager, o *tiled.Object) {
	fmt.Println("Found a text!", o.Text)
	e := em.NewEntity("text")
	img, _ := ebiten.NewImage(100, 100, ebiten.FilterDefault)
	ebitenutil.DebugPrint(img, o.Text.Text)
	if !headless {
		em.Add(e, components.Pos{Vec: gfx.V(o.X, o.Y)})

		conditional := o.Properties.GetString("conditional")
		if conditional != "" {
			maxTransitions := 0 //Default, infinite number of transitions
			i, err := strconv.Atoi(o.Properties.GetString("max_transitions"))
			if err == nil {
				maxTransitions = i
				fmt.Println("loaded", maxTransitions, conditional)
			}

			em.Add(e, components.ConditionalDrawable{ConditionName: conditional, MaxTransitions: maxTransitions})
		}

	}
	em.Add(e, components.Drawable{img})
}

func Path(em *entity.Manager, o *tiled.Object) {
	pathID := em.NewEntity("path")
	center := gfx.V(o.X, o.Y).AddXY(o.Width/2, o.Height/2)
	em.Add(pathID, components.Path{
		Label:  o.Name,
		Points: gfx.Polygon{center, center.AddXY(0, -o.Height/2)},
		Type:   pathanimation.Ellipse,
	})
}

func Animation(em *entity.Manager, pos gfx.Vec, direction float64) {
	img, err := gfx.OpenPNG("assets/images/frames.png")
	if err != nil {
		log.Fatal(err)
	}

	eImg, _ := ebiten.NewImageFromImage(img, ebiten.FilterDefault)

	animID := em.NewEntity("animation")
	// em.Add(animID, components.Following{
	// 	ID:     "player_1",
	// 	Offset: gfx.V(-22, -22),
	// })
	anim := animation.New(eImg, 64, 64)
	anim.Easing = ease.OutCubic
	anim.Direction = direction
	em.Add(animID, anim)
	em.Add(animID, components.Drawable{})
	em.Add(animID, components.Pos{pos.Add(gfx.V(-22, -22))})
}
